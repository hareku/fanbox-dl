# fanbox-dl: Pixiv FANBOX Downloader

**fanbox-dl** downloads all images of a creator. Of course, only the images you can view (supporting or free post).

## Installation

Please download from https://github.com/hareku/fanbox-dl/releases.

- Windows (64bit): `fanbox-dl_x.x.x_Windows_x86_64.exe`
- Windows (32bit): `fanbox-dl_x.x.x_Windows_i386.exe`
- Mac: `fanbox-dl_x.x.x_Darwin_x86_64`

## Usage

See usage `fanbox-dl --help`.

```
NAME:
   fanbox-dl - Downloads all posted original images of the specified user.

USAGE:
   fanbox-dl.out [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --user value      Pixiv user ID to download, don't prepend '@'.
   --sessid value    FANBOXSESSID which is stored in Cookies.
   --save-dir value  Directory for save images. (default: "./images")
   --dir-by-post     Whether to separate save directories for each post. (default: false)
   --all             Whether to check all posts. If --all=false, finish to download when found already downloaded image. (default: false)
   --dry-run         Whether to dry-run (not download images). (default: false)
   --help, -h        show help (default: false)
```

### Example

The case if you want to download all images of `https://www.fanbox.cc/@example`, execute `fanbox-dl --sessid xxxxx --save-dir ./images --user example`.

And you can see images e.g. `./images/example/xxxx.jpg`.

### --sessid (FANBOXSESSID)

fanbox-dl needs FANBOXSESSID which is stored in browser Cookies for login state.

For example, if you are using Google Chrome, you can get it by following the steps in https://developers.google.com/web/tools/chrome-devtools/storage/cookies.
