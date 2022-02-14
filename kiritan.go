package kiritan_handler

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/TKMAX777/winapi"
	"github.com/lxn/win"
)

const windowName = "VOICEROID＋ 東北きりたん EX"

type Handler struct {
	hwnd win.HWND

	controls struct {
		text win.HWND

		play win.HWND
		stop win.HWND
		save win.HWND
		time win.HWND
	}
}

func New() (*Handler, error) {
	var handler = Handler{}
	handler.hwnd = winapi.FindWindow(nil, winapi.MustUTF16PtrFromString(windowName))
	if handler.hwnd == win.HWND(winapi.NULL) {
		handler.hwnd = winapi.FindWindow(nil, winapi.MustUTF16PtrFromString(windowName+"*"))
		if handler.hwnd == win.HWND(winapi.NULL) {
			return nil, errors.New("GetHandler")
		}
	}

	var err error
	func() {
		defer func() {
			var derr = recover()
			if derr != nil {
				err = errors.Errorf("GetControls: %v", derr)
			}
		}()

		var hwnds = winapi.EnumChildWindows(handler.hwnd)
		hwnds = winapi.EnumChildWindows(hwnds[0])
		hwnds = winapi.EnumChildWindows(hwnds[1])
		hwnds = winapi.EnumChildWindows(hwnds[0])
		hwnds = winapi.EnumChildWindows(hwnds[0])
		hwnds = winapi.EnumChildWindows(hwnds[0])
		hwnds = winapi.EnumChildWindows(hwnds[0])

		handler.controls.text = win.HWND(winapi.EnumChildWindows(hwnds[0])[0])
		handler.controls.play = win.HWND(winapi.EnumChildWindows(hwnds[1])[0])
		handler.controls.stop = win.HWND(winapi.EnumChildWindows(hwnds[1])[1])
		handler.controls.save = win.HWND(winapi.EnumChildWindows(hwnds[1])[2])
		handler.controls.time = win.HWND(winapi.EnumChildWindows(hwnds[1])[3])
	}()

	return &handler, err
}

func (h Handler) SetText(text string) error {
	value, err := syscall.UTF16FromString(text)
	if err != nil {
		return errors.Wrap(err, "NewUTF16Pointer")
	}

	var resRow = winapi.SendMessage(
		h.controls.text,
		win.WM_SETTEXT,
		winapi.NULL,
		uintptr(unsafe.Pointer(&value[0])),
	)

	var res = int64(resRow)

	switch res {
	default:
		return errors.New("BoxEditingError")
	case win.LB_ERRSPACE:
		return errors.New("BoxInsufficientSpace")
	}
}

func (h Handler) Play() {
	winapi.SendMessage(
		h.controls.play,
		win.BM_CLICK,
		winapi.NULL,
		winapi.NULL,
	)
}

func (h Handler) Stop() {
	winapi.SendMessage(
		h.controls.stop,
		win.BM_CLICK,
		winapi.NULL,
		winapi.NULL,
	)
}

// Save saves the audio and text file to the specified location
//  This method will overwrite files if they already exist.
func (h Handler) Save(FilePath string) error {
	// press button: "音声保存"
	// this function blocks until the saving dialog destroyed
	go winapi.SendMessage(
		h.controls.save,
		win.BM_CLICK,
		winapi.NULL,
		winapi.NULL,
	)

	var err error

	// if the file already exists, remove it beforehand
	_, err = os.Stat(FilePath)
	if err == nil {
		os.Remove(FilePath)
	}

	var textFilePath = strings.TrimSuffix(FilePath, filepath.Ext(FilePath)) + ".txt"
	_, err = os.Stat(textFilePath)
	if err == nil {
		os.Remove(textFilePath)
	}

	var mainHwnd, editHWND, saveHWND win.HWND
	var ferr = func() error {
		defer func() {
			var derr = recover()
			if derr != nil {
				err = errors.Errorf("GetPtr: %v", derr)
			}
		}()

		for {
			mainHwnd = winapi.FindWindow(winapi.MustUTF16PtrFromString("#32770"), winapi.MustUTF16PtrFromString("音声ファイルの保存"))
			if mainHwnd == win.HWND(winapi.NULL) {
				time.Sleep(time.Millisecond * 500)
				continue
			}
			break
		}

		time.Sleep(time.Second)

		editHWND = winapi.FindWindowEx(mainHwnd, win.HWND(winapi.NULL), winapi.MustUTF16PtrFromString("DUIViewWndClassName"), nil)
		if editHWND == win.HWND(winapi.NULL) {
			return errors.New("NotFoundDUIViewWndClassName")
		}

		editHWND = winapi.FindWindowEx(editHWND, win.HWND(winapi.NULL), winapi.MustUTF16PtrFromString("DirectUIHWND"), nil)
		if editHWND == win.HWND(winapi.NULL) {
			return errors.New("NotFoundDirectUIHWND")
		}

		editHWND = winapi.FindWindowEx(editHWND, win.HWND(winapi.NULL), winapi.MustUTF16PtrFromString("FloatNotifySink"), nil)
		if editHWND == win.HWND(winapi.NULL) {
			return errors.New("NotFoundFloatNotifySink")
		}

		editHWND = winapi.FindWindowEx(editHWND, win.HWND(winapi.NULL), winapi.MustUTF16PtrFromString("ComboBox"), nil)
		if editHWND == win.HWND(winapi.NULL) {
			return errors.New("NotFoundComboBox")
		}

		editHWND = winapi.FindWindowEx(editHWND, win.HWND(winapi.NULL), winapi.MustUTF16PtrFromString("Edit"), nil)
		if editHWND == win.HWND(winapi.NULL) {
			return errors.New("NotFoundEdit")
		}

		saveHWND = winapi.FindWindowEx(mainHwnd, win.HWND(winapi.NULL), winapi.MustUTF16PtrFromString("Button"), nil)
		if saveHWND == win.HWND(winapi.NULL) {
			return errors.New("NotFoundSaveButton")
		}

		winapi.SendMessage(editHWND, win.WM_SETTEXT, winapi.NULL, uintptr(unsafe.Pointer(winapi.MustUTF16PtrFromString(FilePath))))
		go winapi.SendMessage(saveHWND, win.BM_CLICK, winapi.NULL, winapi.NULL)

		time.Sleep(time.Second)

		var dialog win.HWND
		for {
			// search overwrite dialogs
			dialog = winapi.FindWindowEx(win.HWND(winapi.NULL), dialog, winapi.MustUTF16PtrFromString("#32770"), winapi.MustUTF16PtrFromString("音声ファイルの保存"))
			if dialog == win.HWND(winapi.NULL) {
				return nil
			}

			// button: "はい"
			var button = winapi.FindWindowEx(dialog, win.HWND(winapi.NULL), winapi.MustUTF16PtrFromString("Button"), nil)
			if button == win.HWND(winapi.NULL) {
				continue
			}

			go winapi.SendMessage(button, win.BM_CLICK, winapi.NULL, winapi.NULL)
			return nil
		}
	}()

	winapi.SendMessage(mainHwnd, win.WM_CLOSE, winapi.NULL, winapi.NULL)

	switch {
	case err != nil:
		return errors.Wrap(err, "ControlConsoleWindow")
	case ferr != nil:
		return errors.Wrap(ferr, "ControlConsoleWindow")
	}

	return nil
}

func (h Handler) TimeMessage() {
	winapi.SendMessage(
		h.controls.time,
		win.BM_CLICK,
		winapi.NULL,
		winapi.NULL,
	)
}
