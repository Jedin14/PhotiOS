# PhotiOS (Google Photos for iOS) — Unlimited Storage

![demo](readme_assets/app_demo.webp)

Desktop Client for Google Photos with automatic iPhone USB detection, Apple Live Photos pairing, and 0-quota unlimited original quality backup.

> *Open Source notice*: PhotiOS is based on and extends the open-source [gotohp](https://github.com/xob0t/gotohp) engine under the MIT License.

- **Apple iPhone USB Auto-Detection**: Automatically detects your connected iPhone, scans for unsynced photos & videos, and stages them for upload
- **Native HEIC & MOV Previews**: High-performance thumbnail generation and full lightbox preview for Apple media
- **Album Organization**: Upload directly or organize into New Albums, Existing Albums, or Auto-Monthly albums
- **Apple Live Photos**: Pairs `.HEIC` + `.MOV` files into native Google Photos Live Photos
- **Unlimited uploads**: Spoofs original Pixel XL to upload in full Original Quality with 0 bytes Google Drive quota used (can be toggled)
- **Emil Kowalski-inspired UX**: Tactile micro-interactions, smooth cubic-bezier easing, and glassmorphic aesthetics
- Drag-and-drop file upload interface
- Credential management
- Real-time upload progress tracking
- Configurable upload threads
- Individual files or directories uploads, with optional recursive scanning
- Skips files already present in your account
- Configurable, persistent upload settings

> 📖 **Full Documentation**: See [USER_GUIDE.md](USER_GUIDE.md) for step-by-step setup, album selection guide, and zero-quota details.

## [Download](https://github.com/xob0t/gotohp/releases/latest)

## CLI usage

Download the `gotohp-cli` release artifact for your platform. It runs without the GUI.

```shell
gotohp-cli upload /path/to/photos --recursive --threads 5
gotohp-cli creds add "<oauth_token cookie value>"   # Embedded Setup sign-in, see below
gotohp-cli creds add "androidId=..."                 # or a raw credential string
gotohp-cli creds add - < token.txt                   # or read either from stdin
gotohp-cli creds set user@gmail.com
```

The CLI shares the config file with the GUI, but reads only the stored credentials from it. Everything else (proxy, quality, upload behaviour) is set by flags, so a scripted run is unaffected by preferences changed in the GUI. Pass `--config <path>` to use a different config file and `--account <email>` to upload with a specific stored account.

Run `gotohp-cli help` or `gotohp-cli upload --help` for all commands and options.

The GUI binary runs the same CLI when given a command, e.g. `gotohp upload ...`.

## Sign in

### Option 1 - Embedded setup

1. Open gotohp and click **Add Google account**.
2. Click **Open Google sign-in** and sign in with the account you want to add.
3. Sign in on the Google Embedded Setup page and click **I agree**. The page may keep loading forever; this is expected.
4. Open the browser developer tools. Under Application or Storage, open Cookies for `accounts.google.com`.
5. Copy only the value of the `oauth_token` cookie and paste it into gotohp.
6. Click **Connect account**.

From the CLI, open the same page yourself and run `gotohp-cli creds add "<oauth_token value>"`.

### Option 2 - ReVanced. No root required

<details>
  <summary><strong>Click to expand</strong></summary>

1. Install Google Photos ReVanced on your android device/emulator.
   - Install GmsCore [https://github.com/ReVanced/GmsCore/releases](https://github.com/ReVanced/GmsCore/releases)
   - Install patched apk [https://github.com/RookieEnough/Morphe-AutoBuilds/releases](https://github.com/RookieEnough/Morphe-AutoBuilds/releases) or patch it yourself
2. Connect the device to your PC via ADB.
3. Open the terminal on your PC and execute

   Windows

   ```cmd
   adb logcat | FINDSTR "auth%2Fphotos.native"
   ```

   Linux/Mac

   ```shell
   adb logcat | grep "auth%2Fphotos.native"
   ```

4. If you are already using ReVanced - remove Google Account from GmsCore.
5. Open Google Photos ReVanced on your device and log into your account.
6. One or more identical GmsCore logs should appear in the terminal.
7. Copy text from `androidId=` to the end of the line from any log.
8. That's it! 🎉

</details>

### Option 3 - Official apk. Root required

<details>
  <summary><strong>Click to expand</strong></summary>

1. Get a rooted android device or an emulator.
2. Connect the device to your PC via ADB.
3. Install [HTTP Toolkit](https://httptoolkit.com)
4. In HTTP Toolkit, select Intercept - `Android Device via ADB`. Filter traffic with

   ```text
   contains(https://www.googleapis.com/auth/photos.native)
   ```

   Or if you have an older version of Google Photos, try

   ```text
   contains(www.googleapis.com%2Fauth%2Fplus.photos.readwrite)
   ```

5. Open Google Photos app and login with your account.
6. A single request should appear.  
   Copy request body as text.
7. Add that credential string in gotohp.
8. If gotohp asks for a token binding key, keep the rooted device connected and click `Read from ADB`. gotohp will read the account's `lstBindingKeyAlias` from Android AccountManager and save it into the credential.

#### Troubleshooting

- **No Auth Request Intercepted**
  1. Log out of your Google account.
  2. Log in again.
  3. Try `Android App via Frida` interception method in HTTP Toolkit.
- **Token binding key not found**
  1. Make sure the same Google account is present on the connected device.
  2. Make sure root is available to ADB.

</details>

## Build

Follow official wails3 guide
[https://v3.wails.io/getting-started/installation/](https://v3.wails.io/getting-started/installation/)

## Support the Project

If PhotiOS saves you storage space and time, consider supporting development:

[![Support via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=for-the-badge&logo=paypal&logoColor=white)](https://paypal.me/jedinmathew)

* PayPal: **[paypal.me/jedinmathew](https://paypal.me/jedinmathew)**

---

## License & Attribution

PhotiOS is licensed under the [MIT License](LICENSE).  
PhotiOS incorporates, builds upon, and extends the open-source [gotohp](https://github.com/xob0t/gotohp) project by [xob0t](https://github.com/xob0t), used under the terms of the MIT License. Special thanks to the gotohp contributors.
