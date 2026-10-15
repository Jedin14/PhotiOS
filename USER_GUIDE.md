# PhotiOS (Google Photos for iOS) — Unlimited Storage

An Apple-styled desktop companion for **Google Photos** with **automatic iPhone detection**, **Live Photos pairing**, **album organization**, and **unlimited Original Quality backup (0 bytes Google Drive quota used)**.

> *Open Source notice*: PhotiOS is based on and extends the open-source [gotohp](https://github.com/xob0t/gotohp) engine under the MIT License.

---

## Table of Contents
1. [Quick Start (Running the App)](#1-quick-start)
2. [iPhone Auto-Detection & Workflow](#2-iphone-auto-detection--workflow)
3. [Photo Grid & Previews](#3-photo-grid--previews)
4. [Upload Options & Album Selection](#4-upload-options--album-selection)
5. [Unlimited Storage & Zero Quota Explained](#5-unlimited-storage--zero-quota-explained)
6. [Apple Live Photos Support](#6-apple-live-photos-support)
7. [Settings Reference](#7-settings-reference)
8. [Connecting Your Google Account](#8-connecting-your-google-account)
9. [Troubleshooting & FAQ](#9-troubleshooting--faq)

---

## 1. Quick Start

### Standalone Executable
You do not need to install Node.js, Go, or run terminal commands to use the app.
* The pre-built standalone Windows app is located at:
  * **`PhotiOS.exe`** (in this folder)
  * Or **`app.exe`** / `bin/PhotiOS.exe`
* Simply **double-click `PhotiOS.exe`** to launch it. It runs as a native Windows GUI app with no background terminal window popping up.

---

## 2. iPhone Auto-Detection & Workflow

PhotiOS automatically monitors your PC's USB ports for an Apple iPhone using Windows Portable Devices (WPD) and Shell COM.

### How it works:
1. **Connect your iPhone** to your PC using a Lightning or USB-C cable.
2. **Unlock your iPhone**. If prompted with *"Trust This Computer?"*, tap **Trust** and enter your passcode.
3. An elegant Apple-styled banner appears in PhotiOS:
   > **iPhone Connected — Apple iPhone**  
   > Ready to check for unsynced photos. `[ Scan Now ]`
4. Click **Scan Now**:
   * The app traverses the internal camera DCIM storage (e.g. `202609_a`, `202610_a`).
   * It calculates SHA-1 checksums for your photos and videos.
   * It queries Google Photos in real-time to find which files already exist in your cloud library.
   * Only photos and videos that **have not been uploaded yet** are presented for upload!

---

## 3. Photo Grid & Previews

Once the scan completes, you are taken to the **Photo Grid**:

* **Real Thumbnails for HEIC & MOV**: Uses native Windows Shell codecs to extract crisp 320px JPEG thumbnails for all iPhone `.HEIC` photos and `.MOV` videos in ~50ms, cached in `.thumbs` for instant subsequent loads.
* **Video Badges**: Videos are marked with a `MOV` play badge in the corner.
* **Selection Controls**:
  * Click any photo to toggle its selection checkbox.
  * Use **Select All / Deselect All** to toggle all photos at once.
  * Selection count badge tracks selected items dynamically.
* **Fullscreen Lightbox Preview**:
  * **Double-click** any photo or click the **expand icon** (⤢) on hover.
  * High-resolution preview with camera metadata (size, format, file name).
  * Use **Left / Right Arrow keys** (or navigation buttons) to step through photos.
  * Press **Space** to toggle selection directly from the preview.
  * Press **Escape** (or click the close button) to exit.

---

## 4. Upload Options & Album Selection

When photos are selected, the action bar at the bottom provides two distinct options:

### ☁️ Option A: Upload Only (Direct Upload)
* Click **Upload Only** to upload the selected photos directly into your main Google Photos library without creating or adding them to an album.
* 1-click execution — fast and friction-free.

### 📁 Option B: To Album... (Album Sheet)
Clicking **To Album...** opens an Apple-styled action sheet with three modes:

1. **✨ New Album**:
   * Type a new album name (e.g. *"Summer Trip"*, *"Family"*).
   * Or click one-tap suggestion chips:
     * `iPhone Photos`
     * `October 2026` (current month)
     * `September 2026`
   * Click **Upload to Album** — PhotiOS creates the album in Google Photos and adds all uploaded items to it.
2. **🔗 Existing Album / Key**:
   * Enter the exact title of an existing Google Photos album, or paste its `AF1Qip...` album share key.
   * All uploaded photos will be added into that existing album.
3. **🗂️ Auto by Month**:
   * Photos will be automatically grouped into albums named after your iPhone's camera folders (e.g. `202609_a`, `202610_a`).

---

## 5. Unlimited Storage & Zero Quota Explained

### How 0-Quota Backup Works:
When Google released the original **Google Pixel XL** (2016), Google granted it **lifetime, unlimited original quality backup** that never consumes Google Drive or Google One storage.

* PhotiOS transmits the upload client device profile as `Google Pixel XL` with `Quality: 3` (Original Quality) during the commit phase.
* Google Photos servers process the upload under the Pixel XL promotion:
  * Storage Policy: **Original Quality (Uncompressed)**
  * Quota Charged: **0 Bytes**

### Camera & Lens Metadata are 100% Preserved!
Google Photos separates **Camera Info** from **Upload Client Info**:
* **Camera Details (from photo EXIF)**: Stored and displayed completely intact (e.g., *Apple iPhone 15 Pro, 24mm ƒ/1.78, 48 MP, ISO 64, GPS location*).
* **Backup Details (from upload client)**: Shows *Backed up (Original quality) • Uploaded from Pixel XL*.
* Google Photos displays:
  > **"This item doesn't take up space in your account storage. [Learn more](https://support.google.com/photos?p=how_storage_works)"**

---

## 6. Apple Live Photos Support

When you capture photos on an iPhone, iOS saves both a `.HEIC` image and a companion short `.MOV` video sharing an Apple `Content Identifier`.

* When **Pair Apple Live Photos** is enabled (default), PhotiOS reads the Apple metadata and commits both files together as a single interactive Live Photo.
* In Google Photos (on web and iOS/Android app), you can press and hold to play the photo with motion and sound just like on your iPhone!

---

## 7. Settings Reference

Access settings by clicking the gear icon in the top right:

| Setting | Recommended | Description |
| :--- | :---: | :--- |
| **Upload Threads** | `8` | Number of concurrent network uploads (default 8). |
| **Use Quota** | **OFF** ⚪ | When OFF, uploads spoof Pixel XL for **unlimited 0-quota backup**. Turn ON only if you want uploads to count against your 15 GB quota. |
| **Storage Saver Quality** | **OFF** ⚪ | When OFF, files upload in full **Original Quality**. When ON, files are compressed (spoofing Pixel 2). |
| **Force Upload** | **ON** 🟢 | Re-uploads files even if Google Photos already has a hash match (useful if earlier uploads were made directly from iOS and used quota). |
| **Pair Apple Live Photos** | **ON** 🟢 | Bundles iPhone `.HEIC` + `.MOV` pairs into interactive Live Photos. |
| **Update Existing Photos to Live** | **ON** 🟢 | If a static photo was previously uploaded, this updates it into a Live Photo when its video companion is found. |
| **Disable Unsupported Files Filter** | **OFF** ⚪ | Filters out unsupported files before upload. |
| **Delete From Host After Upload** | **OFF** ⚪ | Keeps local files on your device (leaves iPhone files untouched). |

---

## 8. Connecting Your Google Account

PhotiOS supports three connection methods:

### Method 1: Embedded Setup (Recommended)
1. In PhotiOS, click **Add Google account** → **Open Google sign-in**.
2. Sign in with your Google account.
3. On the confirmation page, press `F12` to open Developer Tools.
4. Go to **Application** (or **Storage**) → **Cookies** → `accounts.google.com`.
5. Copy the value of the `oauth_token` cookie and paste it into PhotiOS.
6. Click **Connect account**.

### Method 2: Google Photos ReVanced (No root required)
1. Install GmsCore and Google Photos ReVanced on an Android device or emulator.
2. Run `adb logcat | FINDSTR "auth%2Fphotos.native"`.
3. Copy the string starting with `androidId=` into PhotiOS.

---

## 9. Troubleshooting & FAQ

#### Q: My iPhone is plugged in but the app doesn't detect it.
1. **Unlock your iPhone** with Face ID / passcode.
2. If prompted with *"Trust This Computer?"*, tap **Trust**.
3. Ensure the Apple Mobile Device USB Driver is working (iTunes or Apple Devices app from Microsoft Store can help ensure drivers are installed).
4. Re-plug the USB cable.

#### Q: How can I verify that an uploaded photo used 0 bytes of storage?
1. Open [photos.google.com](https://photos.google.com).
2. Click any photo and tap the **ⓘ Info** icon.
3. Under the details, check that it says:
   `"This item doesn't take up space in your account storage."`
   and `"Uploaded from Pixel XL"`.
4. Check your Google One storage quota at [one.google.com/storage](https://one.google.com/storage) — the Google Photos usage will not increase.

#### Q: Where are cached iPhone files stored?
Temporary scan caches and thumbnails are stored in:  
`%LOCALAPPDATA%\gotohp\iphone_cache\`  
You can delete this directory at any time; `gotohp` will re-create it as needed.

---

## 10. Support the Project

If you love having unlimited 0-quota backup for your iPhone photos, please consider supporting the project:

[![Support via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=for-the-badge&logo=paypal&logoColor=white)](https://paypal.me/jedinmathew)

* Direct link: **[paypal.me/jedinmathew](https://paypal.me/jedinmathew)**
