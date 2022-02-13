# KIRITAN Controller
## 概要
東北きりたんのGUIの一部をGoから制御するライブラリ。

## 仕組み
Win32APIで適当にGUIを操作しているだけ。

## Useage

```go
import "github.com/TKMAX777/kiritan_handler"

func main() {
    kiritan, err := kiritan_handler.New()
    if err != nil {
        panic(err)
    }

    kiritan.SetText("ほげほげ")
    kiritan.Play()
}

```