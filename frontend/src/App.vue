<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { TooltipProvider } from '@/components/ui/tooltip'
import {
  Sheet,
  SheetContent,
  SheetTrigger,
} from '@/components/ui/sheet'
import { useColorMode } from '@vueuse/core'
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { UserPlus, Sparkles, Settings as SettingsIcon, Image as ImageIcon } from '@lucide/vue'
import { ConfigManager } from '../bindings/app/core'
import { Events } from '@wailsio/runtime'
import Button from "./components/ui/button/Button.vue"
import GoogleAccountSelect from './components/GoogleAccountSelect.vue'
import GoogleAuthSetup from "./components/GoogleAuthSetup.vue"
import './index.css'
import SettingsPanel from "./SettingsPanel.vue"
import Upload from './Upload.vue'
import { uploadManager } from './utils/UploadManager'
import Toaster from './components/ui/sonner/Sonner.vue'
import IPhoneBanner from './components/IPhoneBanner.vue'
import IPhonePhotoGrid, { type ScanResult } from './components/IPhonePhotoGrid.vue'
import IPhoneScanProgress from './components/IPhoneScanProgress.vue'
import SupportButton from './components/SupportButton.vue'

import { toast } from "vue-sonner"

useColorMode().value = "dark"

const { state: uploadState } = uploadManager
const copyButtonText = ref('Copy as JSON');

// Drag state for dual drop zones
const isDraggingFiles = ref(false)

// Album upload flow state
const showAlbumInput = ref(false)
const pendingFiles = ref<string[]>([])
const pendingFileCount = ref(0)

const selectedOption = ref('')
const options = ref<string[]>([])
const accountNeedsTokenBinding = ref<Record<string, boolean>>({})
const albumNameOrKey = ref('')
const tokenBindingEmail = ref('')
const isExtractingTokenBinding = ref(false)
const isAccountSetupOpen = ref(false)
const removingAccount = ref('')

interface IPhoneStatus {
  connected: boolean
  deviceName: string
  devicePath: string
}

interface IPhoneScanProgressData {
  phase: string
  total: number
  scanned: number
  unsyncedCount: number
}

// iPhone state
const iphoneStatus = ref<IPhoneStatus | null>(null)
const isIPhoneBannerDismissed = ref(false)
const isScanningIPhone = ref(false)
const iphoneScanProgress = ref<IPhoneScanProgressData | null>(null)
const iphoneScanResult = ref<ScanResult | null>(null)
const currentView = ref<'default' | 'iphone-scanning' | 'iphone-grid'>('default')

function startIPhoneScan() {
  if (options.value.length === 0) {
    toast.error('Connect Google Photos first', {
      description: 'Please add a Google account to check for unsynced photos.',
    })
    openAccountSetup()
    return
  }
  isScanningIPhone.value = true
  iphoneScanProgress.value = { phase: 'discovering', total: 0, scanned: 0, unsyncedCount: 0 }
  currentView.value = 'iphone-scanning'
  Events.Emit('iphone-scan-request')
}

function cancelIPhoneScan() {
  isScanningIPhone.value = false
  currentView.value = 'default'
  Events.Emit('iphone-scan-cancel')
  toast('Scan cancelled')
}

function dismissIPhoneBanner() {
  isIPhoneBannerDismissed.value = true
}

function handleBackFromGrid() {
  currentView.value = 'default'
}

interface IPhoneUploadOptions {
  paths: string[]
  albumMode: 'none' | 'name' | 'auto'
  albumName?: string
}

async function handleIPhoneUpload(opts: IPhoneUploadOptions) {
  try {
    if (opts.albumMode === 'name' && opts.albumName) {
      await ConfigManager.SetAlbumName(opts.albumName)
      await ConfigManager.SetAlbumAutoMode(false)
    } else if (opts.albumMode === 'auto') {
      await ConfigManager.SetAlbumName('')
      await ConfigManager.SetAlbumAutoMode(true)
    } else {
      await ConfigManager.SetAlbumName('')
      await ConfigManager.SetAlbumAutoMode(false)
    }

    Events.Emit('startIPhoneUpload', { files: opts.paths })
    currentView.value = 'default'
  } catch (error) {
    console.error('Failed to configure album and start upload:', error)
    toast.error('Failed to start upload', {
      description: error instanceof Error ? error.message : String(error),
    })
  }
}

function handleIPhoneUploadSelected(paths: string[]) {
  handleIPhoneUpload({ paths, albumMode: 'none' })
}

function handleIPhoneUploadAll() {
  handleIPhoneUpload({ paths: [], albumMode: 'none' })
}

watch(selectedOption, async (newValue) => {
  if (newValue) {
    try {
      await ConfigManager.SetSelected(newValue)
      await updateTokenBindingPrompt(newValue)
      console.log('Successfully updated selected value:', newValue)
    } catch (error) {
      console.error('Failed to update selected value:', error)
      toast.error('Failed to update selected account.')
    }
  } else {
    tokenBindingEmail.value = ''
  }
})

async function updateTokenBindingPrompt(email: string) {
  tokenBindingEmail.value = accountNeedsTokenBinding.value[email] ? email : ''
}

async function refreshCredentials() {
  try {
    const state = await ConfigManager.GetAccounts()
    const nextNeedsTokenBinding: Record<string, boolean> = {}
    const nextOptions = state.accounts.map(account => {
      nextNeedsTokenBinding[account.email] = account.needsTokenBinding
      return account.email
    })

    accountNeedsTokenBinding.value = nextNeedsTokenBinding
    options.value = nextOptions
    selectedOption.value = state.selected || ''
    if (state.selected) {
      await updateTokenBindingPrompt(state.selected)
    } else {
      tokenBindingEmail.value = ''
    }
  } catch (error) {
    console.error('Failed to refresh Google accounts:', error)
    toast.error('Could not refresh Google accounts', {
      description: error instanceof Error ? error.message : String(error),
    })
  }
}

async function addTokenBindingAliasFromADB() {
  if (!tokenBindingEmail.value) return

  isExtractingTokenBinding.value = true
  try {
    await ConfigManager.AddTokenBindingAliasFromADB(tokenBindingEmail.value)
    await refreshCredentials()
    tokenBindingEmail.value = ''
    toast.success('Token binding key added.')
  } catch (error) {
    console.error('Failed to add token binding key:', error)
    toast.error('Failed to add token binding key', {
      description: error instanceof Error ? error.message : String(error),
    })
  } finally {
    isExtractingTokenBinding.value = false
  }
}

async function removeCredentials(email: string) {
  removingAccount.value = email
  try {
    await ConfigManager.RemoveCredentials(email)

    const removedSelectedAccount = selectedOption.value === email
    await refreshCredentials()
    if (removedSelectedAccount && options.value.length > 0 && !selectedOption.value) {
      selectedOption.value = options.value[0]
    }
    toast.success('Credentials removed.')
    return true
  } catch (error) {
    console.error('Failed to remove credentials:', error)
    toast.error('Failed to remove credentials.')
    return false
  } finally {
    removingAccount.value = ''
  }
}

function openAccountSetup() {
  isAccountSetupOpen.value = true
}

onMounted(async () => {
  await refreshCredentials()

})

const handleCopyClick = () => {
  uploadManager.copyResultsAsJson();
  copyButtonText.value = 'Copied!';
  setTimeout(() => copyButtonText.value = 'Copy as JSON', 1000);
};



// Global drag event handlers to detect file dragging
let dragLeaveTimeout: ReturnType<typeof setTimeout> | null = null

const onDragEnter = (e: DragEvent) => {
  if (!e.dataTransfer?.types.includes('Files')) return
  
  // Clear any pending drag leave timeout
  if (dragLeaveTimeout) {
    clearTimeout(dragLeaveTimeout)
    dragLeaveTimeout = null
  }
  
  isDraggingFiles.value = true
}

const onDragOver = (e: DragEvent) => {
  if (!e.dataTransfer?.types.includes('Files')) return
  e.preventDefault()
  
  // Clear any pending drag leave timeout - we're still dragging
  if (dragLeaveTimeout) {
    clearTimeout(dragLeaveTimeout)
    dragLeaveTimeout = null
  }
}

const onDragLeave = (e: DragEvent) => {
  if (!e.dataTransfer?.types.includes('Files')) return
  
  // Use timeout to detect if we've truly left the window
  // dragover will cancel this if we're still in the window
  if (dragLeaveTimeout) {
    clearTimeout(dragLeaveTimeout)
  }
  dragLeaveTimeout = setTimeout(() => {
    isDraggingFiles.value = false
    dragLeaveTimeout = null
  }, 50)
}

const onDrop = () => {
  // Clear any pending timeout
  if (dragLeaveTimeout) {
    clearTimeout(dragLeaveTimeout)
    dragLeaveTimeout = null
  }
  
  // Delay resetting isDraggingFiles to allow Wails to process the drop target
  // before Vue re-renders and hides the drop zones
  setTimeout(() => {
    isDraggingFiles.value = false
  }, 100)
}

// Album upload confirmation
const confirmAlbumUpload = async () => {
  // Set album name in backend (not persisted to disk)
  await ConfigManager.SetAlbumName(albumNameOrKey.value)
  await ConfigManager.SetAlbumAutoMode(false)
  // Start upload with pending files
  Events.Emit('startUpload', { files: pendingFiles.value })
  showAlbumInput.value = false
  pendingFiles.value = []
  pendingFileCount.value = 0
  albumNameOrKey.value = '' // Reset for next upload
}

const cancelAlbumUpload = () => {
  showAlbumInput.value = false
  pendingFiles.value = []
  pendingFileCount.value = 0
  albumNameOrKey.value = '' // Reset on cancel too
}

// Handle album error event
const albumErrorHandler = (e: Event) => {
  const event = e as CustomEvent<{ AlbumName: string; Error: string }>
  const { AlbumName, Error } = event.detail
  // Check if it's a 404 error (album key not found)
  if (Error.includes('404')) {
    toast.error('Album not found', {
      description: `The album key "${AlbumName}" does not exist or is invalid.`,
    })
  } else {
    toast.error('Failed to create album', {
      description: `Album "${AlbumName}": ${Error}`,
    })
  }
}

const uploadErrorHandler = (e: Event) => {
  const event = e as CustomEvent<{ FileName: string; Message: string }>
  const { FileName, Message } = event.detail
  const errorMessage = Message.replace(/^Error:\s*/, '')
  toast.error(FileName ? `Upload failed: ${FileName}` : 'Upload failed', {
    description: errorMessage,
    duration: 10000,
    important: true,
  })
}

onMounted(() => {
  document.addEventListener('dragenter', onDragEnter)
  document.addEventListener('dragleave', onDragLeave)
  document.addEventListener('dragover', onDragOver)
  document.addEventListener('drop', onDrop)
  window.addEventListener('albumError', albumErrorHandler)
  window.addEventListener('uploadError', uploadErrorHandler)

  // Listen for files-dropped event from backend
  Events.On('files-dropped', async (event: { data: { files: string[]; dropZone: string } }) => {
    const { files, dropZone } = event.data

    if (dropZone === 'album') {
      // Show album input screen
      pendingFiles.value = files
      pendingFileCount.value = files.length
      showAlbumInput.value = true
    } else {
      try {
        await ConfigManager.SetAlbumName('')
        await ConfigManager.SetAlbumAutoMode(dropZone === 'auto-album')
        await Events.Emit('startUpload', { files })
      } catch (error) {
        console.error('Failed to start upload:', error)
        toast.error('Failed to start upload', {
          description: error instanceof Error ? error.message : String(error),
        })
      }
    }
  })

  // iPhone connection events
  Events.On('iphone-connected', (event: { data: IPhoneStatus }) => {
    iphoneStatus.value = event.data
    isIPhoneBannerDismissed.value = false
    toast.success(`iPhone Connected: ${event.data.deviceName}`, {
      description: 'Ready to check for unsynced photos.',
      action: {
        label: 'Scan Now',
        onClick: () => startIPhoneScan(),
      },
    })
  })

  Events.On('iphone-disconnected', () => {
    if (iphoneStatus.value?.connected) {
      toast('iPhone Disconnected')
    }
    iphoneStatus.value = null
    isScanningIPhone.value = false
    if (currentView.value === 'iphone-scanning' || currentView.value === 'iphone-grid') {
      currentView.value = 'default'
    }
  })

  Events.On('iphone-scan-progress', (event: { data: IPhoneScanProgressData }) => {
    iphoneScanProgress.value = event.data
  })

  Events.On('iphone-scan-complete', (event: { data: ScanResult }) => {
    isScanningIPhone.value = false
    iphoneScanResult.value = event.data
    if (event.data.unsyncedCount > 0) {
      currentView.value = 'iphone-grid'
      toast.success('Scan Complete', {
        description: `Found ${event.data.unsyncedCount} unsynced photo${event.data.unsyncedCount !== 1 ? 's' : ''} to upload.`,
      })
    } else {
      currentView.value = 'default'
      toast.success('All Synced!', {
        description: `All ${event.data.totalOnPhone} photos on your iPhone are already in Google Photos.`,
      })
    }
  })
})

onUnmounted(() => {
  document.removeEventListener('dragenter', onDragEnter)
  document.removeEventListener('dragleave', onDragLeave)
  document.removeEventListener('dragover', onDragOver)
  document.removeEventListener('drop', onDrop)
  window.removeEventListener('albumError', albumErrorHandler)
  window.removeEventListener('uploadError', uploadErrorHandler)
  if (dragLeaveTimeout) {
    clearTimeout(dragLeaveTimeout)
  }
})
</script>

<template>
  <main
    class="w-screen h-screen flex flex-col items-center"
    style="--wails-draggable: drag"
  >
    <!-- Drop zones shown when dragging files -->
    <div
      v-if="!uploadState.isUploading && isDraggingFiles && options.length > 0"
      class="w-screen h-screen flex flex-col gap-3 p-6"
      style="--wails-draggable: none"
    >
      <div
        data-file-drop-target
        data-drop-zone="regular"
        class="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-muted-foreground/50 rounded-xl transition-all duration-200 drop-zone"
      >
        <h2 class="text-xl font-semibold select-none text-muted-foreground">
          Upload Only
        </h2>
        <p class="text-sm text-muted-foreground/70 mt-2 select-none text-center px-4">
          Upload files without adding to any album
        </p>
      </div>
      <div
        data-file-drop-target
        data-drop-zone="album"
        class="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-muted-foreground/50 rounded-xl transition-all duration-200 drop-zone"
      >
        <h2 class="text-xl font-semibold select-none text-muted-foreground">
          Upload to Album
        </h2>
        <p class="text-sm text-muted-foreground/70 mt-2 select-none text-center px-4">
          Upload and add to a specific album (you'll enter the name)
        </p>
      </div>
      <div
        data-file-drop-target
        data-drop-zone="auto-album"
        class="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-muted-foreground/50 rounded-xl transition-all duration-200 drop-zone"
      >
        <h2 class="text-xl font-semibold select-none text-muted-foreground">
          Auto Album
        </h2>
        <p class="text-sm text-muted-foreground/70 mt-2 select-none text-center px-4">
          Upload and create albums automatically based on folder names
        </p>
      </div>
    </div>

    <!-- iPhone photo grid view -->
    <div
      v-else-if="!uploadState.isUploading && currentView === 'iphone-grid' && iphoneScanResult"
      class="w-screen h-screen flex flex-col"
      style="--wails-draggable: none"
    >
      <IPhonePhotoGrid
        :scan-result="iphoneScanResult"
        :is-uploading="uploadState.isUploading"
        @upload="handleIPhoneUpload"
        @upload-selected="handleIPhoneUploadSelected"
        @upload-all="handleIPhoneUploadAll"
        @back="handleBackFromGrid"
      />
    </div>

    <!-- iPhone scanning view -->
    <div
      v-else-if="!uploadState.isUploading && currentView === 'iphone-scanning'"
      class="w-screen h-screen flex flex-col items-center justify-center p-6"
      style="--wails-draggable: none"
    >
      <IPhoneScanProgress
        :phase="iphoneScanProgress?.phase || 'discovering'"
        :total="iphoneScanProgress?.total || 0"
        :scanned="iphoneScanProgress?.scanned || 0"
        :unsynced-count="iphoneScanProgress?.unsyncedCount || 0"
        @cancel="cancelIPhoneScan"
      />
    </div>

    <!-- Normal UI (not dragging) -->
    <!-- Normal UI (not dragging) -->
    <div
      v-else-if="!uploadState.isUploading"
      class="w-screen h-screen flex flex-col items-center gap-3.5 max-w-md px-6 justify-between select-none"
      :class="iphoneStatus?.connected && !isIPhoneBannerDismissed ? 'pt-6 pb-2' : 'pt-10 pb-2'"
      data-file-drop-target
    >
      <div class="w-full flex flex-col items-center gap-3.5">
        <!-- PhotiOS Brand Header -->
        <div class="flex flex-col items-center gap-1 text-center select-none brand-header">
          <div class="relative flex items-center justify-center">
            <div class="absolute -inset-1 rounded-2xl bg-gradient-to-r from-emerald-500/20 via-sky-500/20 to-purple-500/20 blur-md opacity-75 animate-pulse" />
            <div class="relative flex size-11 items-center justify-center rounded-2xl bg-gradient-to-br from-zinc-800 to-zinc-900 border border-white/10 shadow-lg">
              <svg class="size-6 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14.5 4h-5L7 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-3l-2.5-3z"/>
                <circle cx="12" cy="13" r="3"/>
              </svg>
            </div>
          </div>
          <div class="flex items-center gap-2 mt-1">
            <h1 class="text-xl font-bold tracking-tight text-foreground">PhotiOS</h1>
            <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
              <Sparkles class="size-3" /> Unlimited
            </span>
          </div>
          <p class="text-[11px] text-muted-foreground/80">
            Google Photos for iOS • Zero Quota Backup
          </p>
        </div>

        <!-- iPhone connection banner -->
        <div
          v-if="iphoneStatus?.connected && !isIPhoneBannerDismissed"
          class="w-full mb-1"
          style="--wails-draggable: none"
        >
          <IPhoneBanner
            :device-name="iphoneStatus.deviceName"
            :is-scanning="isScanningIPhone"
            :scan-progress="iphoneScanProgress || undefined"
            @scan="startIPhoneScan"
            @dismiss="dismissIPhoneBanner"
            @cancel-scan="cancelIPhoneScan"
          />
        </div>
        <template v-if="options.length === 0">
          <div class="flex max-w-xs flex-col items-center gap-4 text-center mt-4">
            <div class="flex size-11 items-center justify-center rounded-full bg-muted text-muted-foreground">
              <UserPlus class="size-5" />
            </div>
            <div class="flex flex-col gap-1">
              <h2 class="text-lg font-semibold select-none">
                Connect Google Photos
              </h2>
              <p class="text-xs text-muted-foreground select-none">
                Add an account before uploading photos and videos.
              </p>
            </div>
            <div class="flex items-center gap-2">
              <Button
                class="cursor-pointer select-none"
                @click="openAccountSetup"
              >
                Add Google account
              </Button>
              <SupportButton variant="compact" show-heart />
            </div>
          </div>
        </template>

        <template v-else>
          <!-- Show album input screen when files dropped on album zone -->
          <template v-if="showAlbumInput">
            <div
              class="flex flex-col items-center justify-center gap-5 p-6"
              style="--wails-draggable: none"
            >
              <h2 class="text-lg font-semibold select-none">
                Upload to Album
              </h2>
              <p class="text-xs text-muted-foreground select-none">
                {{ pendingFileCount }} file(s) ready to upload
              </p>
              
              <div class="flex flex-col gap-2 w-full max-w-xs">
                <Label
                  for="album-input"
                  class="text-muted-foreground text-xs"
                >Album name or key</Label>
                <Input
                  id="album-input"
                  v-model="albumNameOrKey"
                  placeholder="Album name or AF1Qip... key"
                  autofocus
                />
              </div>

              <div class="flex gap-3">
                <Button
                  variant="outline"
                  class="cursor-pointer select-none text-xs"
                  @click="cancelAlbumUpload"
                >
                  Cancel
                </Button>
                <Button
                  class="cursor-pointer select-none text-xs"
                  :disabled="!albumNameOrKey.trim()"
                  @click="confirmAlbumUpload"
                >
                  Upload
                </Button>
              </div>
            </div>
          </template>

          <!-- Normal UI when not dragging -->
          <template v-else>
            <div class="flex flex-col items-center gap-2 w-full max-w-xs">
              <GoogleAccountSelect
                v-model="selectedOption"
                :options="options"
                :removing-account="removingAccount"
                @item-removed="removeCredentials"
                @add="openAccountSetup"
              />
            </div>

            <div
              v-if="tokenBindingEmail"
              class="w-full max-w-xs border rounded-lg p-3 flex flex-col gap-3"
              style="--wails-draggable: none"
            >
              <p class="text-xs text-muted-foreground">
                This credential needs a token binding key from the rooted Android device it was captured from.
              </p>
              <Button
                class="cursor-pointer select-none text-xs"
                :disabled="isExtractingTokenBinding"
                @click="addTokenBindingAliasFromADB"
              >
                {{ isExtractingTokenBinding ? 'Reading ADB...' : 'Read from ADB' }}
              </Button>
            </div>

            <!-- Drop zone indicator card -->
            <div 
              class="w-full max-w-xs py-5 px-4 border border-dashed border-white/10 rounded-xl flex flex-col items-center justify-center gap-2 bg-white/[0.02] hover:bg-white/[0.04] transition-colors select-none"
            >
              <div class="size-8 rounded-full bg-white/5 flex items-center justify-center text-muted-foreground">
                <ImageIcon :size="16" />
              </div>
              <p class="text-xs font-medium text-muted-foreground/90 text-center">
                Drop photos or connect your iPhone to sync
              </p>
            </div>

            <div class="flex items-center gap-2 select-none">
              <Sheet>
                <SheetTrigger>
                  <Button
                    variant="outline"
                    class="cursor-pointer select-none flex items-center gap-1.5 h-8 px-3 text-xs active:scale-95 transition-transform"
                  >
                    <SettingsIcon :size="13" />
                    <span>Settings</span>
                  </Button>
                </SheetTrigger>
                <SheetContent side="bottom">
                  <TooltipProvider disable-hoverable-content>
                    <SettingsPanel />
                  </TooltipProvider>
                </SheetContent>
              </Sheet>

              <SupportButton variant="compact" show-heart />
            </div>

            <div
              v-if="uploadState.uploadedFiles > 0 || uploadState.results.fail.length > 0"
              class="flex flex-col items-center gap-2 border rounded-lg p-4 mt-2 w-full max-w-xs"
            >
              <h3 class="text-sm font-semibold select-none">
                Upload Results
              </h3>
              <div class="text-xs text-muted-foreground flex flex-col gap-0.5">
                <span>Successful: {{ uploadState.results.success.length }}</span>
                <span>Failed: {{ uploadState.results.fail.length }}</span>
                <span>Skipped: {{ uploadState.results.skipped.length }}</span>
                <span>Warnings: {{ uploadState.results.warnings.length }}</span>
              </div>
              <Button
                variant="outline"
                class="cursor-pointer select-none text-xs min-w-[120px]"
                @click="handleCopyClick"
              >
                {{ copyButtonText }}
              </Button>
            </div>
          </template>
        </template>
      </div>

      <!-- Discreet Open Source Attribution Footer -->
      <div class="mt-auto pt-2 pb-1 select-none text-center">
        <p class="text-[10px] text-muted-foreground/35 hover:text-muted-foreground/60 transition-colors">
          PhotiOS • Based on gotohp engine • Open Source (MIT)
        </p>
      </div>
    </div>
    <div
      v-if="uploadState.isUploading"
      class="w-full h-full"
    >
      <Upload />
    </div>
    <GoogleAuthSetup
      v-model:open="isAccountSetupOpen"
      @account-added="refreshCredentials"
    />
    <Toaster
      position="bottom-center"
      rich-colors
      expand
      :visible-toasts="4"
    />
  </main>
</template>
