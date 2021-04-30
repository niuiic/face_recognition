// @Title			main.go
// @Description		gui program to call the face recognition program implemented on the atlas platform.
// @Author 	  		niuiic
// @Update    		niuiic 2021/04/29

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// TODO：完善配置内容

type Config struct {
	PresentServerPath string `json:"present_server_path"`
	PresentServerIp   string `json:"present_server_ip"`
}

func openServer() {
	// TODO：打开presentserver
	println("open server")
}

func readConfig() Config {
	var config Config
	data, err := ioutil.ReadFile("./param.json")
	if err != nil {
		panic("failed to open json config file")
	} else {
		fmt.Println(string(data))
		err = json.Unmarshal(data, &config)
		if err != nil {
			fmt.Println(err)
			panic("failed to parse json data")
		}
	}
	return config
}

func main() {

	config := readConfig()
	faceRecognition := app.New()

	welcomeWindow := faceRecognition.NewWindow("face recognition")
	switchWindow := faceRecognition.NewWindow("switch")
	cameraWindow := faceRecognition.NewWindow("camera")
	localVideoWindow := faceRecognition.NewWindow("local video")

	welcomeWindow.SetContent(
		container.NewVBox(
			widget.NewLabel("Welcome to the face recognition app"),
			widget.NewButton("Go", func() {
				go openServer()
				// welcomeWindow.Close()
				// switchWindow.Show()
				welcomeWindow.SetContent(
					container.NewVBox(
						widget.NewLabel("You can choose local video or camera input to recognition face"),
						widget.NewButton("camera", func() {
							switchWindow.Hide()
							// TODO：打开开发板程序
							cameraWindow.Show()
						}),
						widget.NewButton("local video", func() {
							switchWindow.Hide()
							localVideoWindow.Show()
						}),
					),
				)
			}),
		),
	)

	switchWindow.SetContent(
		container.NewVBox(
			widget.NewLabel("You can choose local video or camera input to recognition face"),
			widget.NewButton("camera", func() {
				switchWindow.Hide()
				// TODO：打开开发板程序
				cameraWindow.Show()
			}),
			widget.NewButton("local video", func() {
				switchWindow.Hide()
				localVideoWindow.Show()
			}),
		),
	)

	cameraWindow.SetContent(
		container.NewVBox(
			widget.NewButton("return", func() {
				cameraWindow.Hide()
				// TODO：关闭摄像机等
				switchWindow.Show()
			}),
			widget.NewButton("click to open the browser",
				func() {
					var commands = map[string]string{
						"windows": "cmd /c start",
						"darwin":  "open",
						"linux":   "xdg-open",
					}
					run, ok := commands[runtime.GOOS]
					if !ok {
						_ = fmt.Errorf("don't know how to open browser on %s platform", runtime.GOOS)
					}

					cmd := exec.Command(run, config.PresentServerIp)
					cmd.Start()
				},
			),
		),
	)

	pathToVideo := widget.NewEntry()
	pathToVideo.SetPlaceHolder("please input the absolute path to your video")
	var label = widget.NewLabel("")

	localVideoContainer :=
		container.NewVBox(
			widget.NewButton("return", func() {
				localVideoWindow.Hide()
				// TODO：关闭摄像机等
				switchWindow.Show()
			}), label,
		)

	form := widget.NewForm(
		widget.NewFormItem("Path", pathToVideo),
	)

	form.OnSubmit = func() {
		var (
			isExits                  bool
			isFile                   bool
			isMP4File                bool
			isAtTheCorrectResolution bool
		)
		cmd := exec.Command("ls", pathToVideo.Text)
		_, err := cmd.Output()
		if err != nil {
			isExits = false
			isFile = false
			isMP4File = false
		} else {
			isExits = true
			s, _ := os.Lstat(pathToVideo.Text)
			isFile = !s.IsDir()
			isMP4File = path.Ext(path.Base(pathToVideo.Text)) == ".mp4"
			if isFile && isMP4File {
				cmd := exec.Command("mplayer", " -identify", " -frames", "5", " -endpos", " 0 ", "-vo", " null", pathToVideo.Text)
				output, err := cmd.Output()
				if err != nil {
					fmt.Println(err)
				} else {
					if strings.Contains(string(output), "ID_VIDEO_WIDTH=1280") && strings.Contains(string(output), "ID_VIDEO_HEIGHT=720") {
						isAtTheCorrectResolution = true
					} else {
						isAtTheCorrectResolution = false
					}
				}
			}
		}
		if !isExits {
			label.SetText("No such file or directory")
		} else if !isFile {
			label.SetText("It's not a file")
		} else if !isMP4File {
			label.SetText("It is not an MP4 file")
		} else if !isAtTheCorrectResolution {
			label.SetText("You video is not at correct resolution. Please set the resolution to 1280 * 720")
		} else {
			label.SetText("Path verification is successful. The video is now being transferred to the development board. Please wait.")
			// TODO：传输视频，显示打开浏览器按钮，启动开发板程序
		}
	}

	localVideoContainer.Add(form)

	localVideoWindow.SetContent(
		localVideoContainer)

	welcomeWindow.Resize(fyne.NewSize(400, 400))
	welcomeWindow.ShowAndRun()
}
