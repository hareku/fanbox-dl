# fanbox-dl: Pixiv FANBOX Downloader

`fanbox-dl` downloads images of supporting and following creators from FANBOX.

## Installation

Please download from https://github.com/hareku/fanbox-dl/releases.

- Windows (64bit): `fanbox-dl_x.x.x_Windows_x86_64.exe`
- Windows (32bit): `fanbox-dl_x.x.x_Windows_i386.exe`
- Mac: `fanbox-dl_x.x.x_Darwin_x86_64`
- Mac (M1 CPU): `fanbox-dl_x.x.x_Darwin_arm64`

## Usage

See usage `fanbox-dl --help`.

### Example

The case if you want to download all images of `https://www.fanbox.cc/@example`, execute `fanbox-dl --sessid xxxxx --save-dir ./images --creator creatornamehere`.

And you can see images e.g. `./images/example/xxxx.jpg`.

### --sessid (FANBOXSESSID)

fanbox-dl needs FANBOXSESSID which is stored in browser Cookies for login state.

For example, if you are using Google Chrome, you can get it by following the steps in https://developers.google.com/web/tools/chrome-devtools/storage/cookies.

## Contribution

Please open an issue or pull request.
