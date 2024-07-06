# fanbox-dl: Pixiv FANBOX Downloader

`fanbox-dl` will download media of supported and followed creators on FANBOX.

Caution: `fanbox-dl` is command-line-program, so it doesn't provide graphical user interface.

## Installation

The latest binary can be downloaded [here](https://github.com/hareku/fanbox-dl/releases/latest).

- Windows (64bit): `fanbox-dl_x.x.x_Windows_x86_64.exe`
- Windows (32bit): `fanbox-dl_x.x.x_Windows_i386.exe`
- Mac: `fanbox-dl_x.x.x_Darwin_x86_64`
- Mac (M1 CPU): `fanbox-dl_x.x.x_Darwin_arm64`

## Usage

1. Open a command line interpreter. For example, If you are Windows user, open `Command Prompt` or `PowerShell`. If you are Mac user, open `Terminal`.

2. Execute the downloaded `fanbox-dl` binary. You can see usage by running `fanbox-dl --help`.

| Command | Description | Usage | Default |
| --- | --- | --- | ---: |
| sessid | Requires FANBOXSESSID which is stored in browser Cookies for login state. <br>When not provided, refers FANBOXSESSID environment value. <br>If unavailable, only free posts are downloaded when accompanied by a `creator` flag. | `--sessid xxxxx` | `NULL` |
| cookie | Cookie string to use for requests. <br>When not provided, refers to the `sessid` flag. | `--cookie "name=value; name2=value2"` | `NULL` |
| creator | Comma separated Pixiv creator IDs to download the contents. <br>Overrides `supporting` and `following` flags. <br>`https://www.fanbox.cc/@`**example**. <br>Only bold text needed from URL. | `--creator user1`, `--creator user1,user2` | `NULL` |
| ignore-creator | Comma separated Pixiv creator IDs to ignore to download the contents. | `--ignore-creator user1,user2` | `NULL` |
| supporting | When disabled, will not download content from creators you're supporting. | `--supporting=false` | `true` |
| following | When disabled, will not download content from creators you only follow. | `--following=false` | `true` |
| dir-by-plan | Separates content saved into directories based on the plan the post belonged to. | `--dir-by-plan` | `false` |
| dir-by-post | Separates content saved into directories based on the title of the post. <br>Stored inside the plan directory when accompanied by the `dir-by-plan` flag. | `--dir-by-post` | `false` |
| all | Will ensure that all content is downloaded from creators. <br>Will also redownload content that might already be present locally. | `--all` | `false` |
| skip-files | Will skip downloading non-image files from creators. | `--skip-files` | `false` |
| dry-run | Will skip downloading all content from creators. | `--dry-run` | `false` |
| verbose | Gives more detailed information about commands being executed by the application. <br>Useful for debugging errors. | `--verbose` | `false` |
| save-dir | Root directory to save content. <br>Put directory in double quotes `"` if it contains spaces. <br> Supports relative and absolute directories. | `--savedir ./content` | `./images` |
| user-agent | User agent to use for requests. | `--user-agent "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"` | `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3` |

### Example

If you want to re-download all images from the creator `https://www.fanbox.cc/@creatornamehere`, execute `fanbox-dl --sessid xxxxx --save-dir ./content --creator creatornamehere --all`.

And you can see media in the relevant directory. `./content/creatornamehere/xxxx.jpg`.

### Acquiring your FANBOXSESSID

fanbox-dl needs your account FANBOXSESSID to download supported content, which has your login state stored in a browser Cookie.

For example, if you are using Google Chrome, you can get it by following the steps in https://developers.google.com/web/tools/chrome-devtools/storage/cookies.

## Contribution

Please open an issue or pull request.
