### Do (Windows only .-.)

### Required Softwares
- git
- less
- rg
- tar

### Build and Minimizing Binary Size
```sh
$ go build -o do.exe -ldflags="-s -w" -trimpath .\cmd\do-go
$ upx --best --lzma do.exe
```

### Memo
- [Color](https://stackoverflow.com/questions/6297072/color-for-the-prompt-just-the-prompt-proper-in-cmd-exe-and-powershell)
