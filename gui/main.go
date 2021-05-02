// @Title : main.go
// @Description : gui program to call the face recognition program implemented on the atlas platform.
// @Author : niuiic
// @Update : 2021/04/29

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"golang.org/x/crypto/ssh"

	"github.com/ThomasRooney/gexpect"
	"github.com/pkg/sftp"
)

// TODO：完善配置内容

// Config : environmental parameter configuration

type Config struct {
	// the path to presenter server
	PresenterServerPath string `json:"presenter_server_path"`
	// IP address monitored by presenter server
	PresenterServerIp string `json:"presenter_server_ip"`
	// port monitored by presenter server
	PresenterServerPort string `json:"presenter_server_port"`
	// output directory of presenter server
	PresenterServerOutputDir string `json:"presenter_server_output_dir"`
	// Ip address of develop board
	DevelopBoardIP string `json:"develop_board_ip"`
	// username of develop board
	DevelopBoardUser string `json:"develop_board_user"`
	// root password of develop board
	DevelopBoardRootPassword string `json:"develop_board_root_password"`
	// the path to face recognition project on develop board
	DevelopBoardProjectPath string `json:"develop_board_project_path"`
}

// open presenter server

func openPresenterServer(config *Config) {
	cmd := "sh " + config.PresenterServerPath
	child, err := gexpect.Spawn(cmd)
	if err != nil {
		log.Fatal("spawn cmd error ", err)
	}

	if err := child.SendLine(config.PresenterServerOutputDir); err != nil {
		log.Fatal("sendLine output dir error ", err)
	}
}

// read the config file

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

func getSshConnect(config *Config) *ssh.Client {
	clientConfig := &ssh.ClientConfig{
		Timeout:         10 * time.Second,
		User:            config.DevelopBoardUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	clientConfig.Auth = []ssh.AuthMethod{ssh.Password(config.DevelopBoardRootPassword)}
	addr := fmt.Sprintf("%s:%d", config.DevelopBoardIP, 22)
	sshClient, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		fmt.Println(err)
	} else {
		return sshClient
	}
	return nil
}

func transferVideo(localVideoPath string, config *Config, sshClient *ssh.Client) {
	sftpClient, err := sftp.NewClient(sshClient)
	defer sftpClient.Close()
	if err != nil {
		log.Fatal(err)
	}

	srcFile, err := os.Open(localVideoPath)
	defer srcFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	remoteFileName := path.Base(localVideoPath)
	dstFile, err := sftpClient.Create(path.Join(config.DevelopBoardProjectPath+`/out`, remoteFileName))
	defer dstFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf)
	}
}

func execFaceRecognition(sshClient *ssh.Client, config *Config, videoName string, exitChan chan struct{}) {
	session, err := sshClient.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()

	cmd := "cd " + config.DevelopBoardProjectPath + " && cd " + `./out` + ` && ./main ` + videoName
	err = session.Start(cmd)

	if err != nil {
		log.Fatal(err)
	}

	<-exitChan
	session2, err := sshClient.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	defer session2.Close()

	output, err := session2.Output(`ps -A | grep main`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))

	PIDRegexp := regexp.MustCompile(`^ ([\d]{4}) ?[\s]+[\d]{2}:[\d]{2}:[\d]{2} main$`)
	PID := PIDRegexp.FindStringSubmatch(string(output))
	for _, pid := range PID {
		println(pid)
	}

}

func main() {

	config := readConfig()
	sshClient := getSshConnect(&config)
	faceRecognition := app.New()
	exitChan := make(chan struct{}, 1)

	mainWindow := faceRecognition.NewWindow("face recognition")

	var (
		welcomePage    *fyne.Container
		switchPage     *fyne.Container
		cameraPage     *fyne.Container
		localVideoPage *fyne.Container
	)

	welcomePage = container.NewVBox(
		widget.NewLabel("Welcome to the face recognition app"),
		widget.NewButton("Go", func() {
			go openPresenterServer(&config)
			mainWindow.SetContent(switchPage)
		}),
	)

	mainWindow.SetContent(welcomePage)

	switchPage = container.NewVBox(
		widget.NewLabel("You can choose local video or camera input to recognition face"),
		widget.NewButton("camera", func() {
			go execFaceRecognition(sshClient, &config, "", exitChan)
			mainWindow.SetContent(cameraPage)
			cameraPage.Show()
		}),
		widget.NewButton("local video", func() {
			mainWindow.SetContent(localVideoPage)
			localVideoPage.Show()
		}),
	)

	cameraPage = container.NewVBox(
		widget.NewButton("return", func() {
			// TODO：关闭摄像机等
			mainWindow.SetContent(switchPage)
			switchPage.Show()
		}),
		widget.NewButton("open browser",
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

				cmd := exec.Command(run, "http://"+config.PresenterServerIp+":"+config.PresenterServerPort)
				err := cmd.Run()
				if err != nil {
					fmt.Println(err)
				}
			},
		),
	)

	btnOpenBrowser := widget.NewButton("open browser",
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

			cmd := exec.Command(run, "http://"+config.PresenterServerIp+":"+config.PresenterServerPort)
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
		},
	)

	pathToVideo := widget.NewEntry()
	pathToVideo.SetPlaceHolder("please input the absolute path to your video")
	var label = widget.NewLabel("")

	localVideoPage =
		container.NewVBox(
			widget.NewButton("return", func() {
				// TODO：关闭摄像机等
				exitChan <- struct{}{}
				mainWindow.SetContent(switchPage)
				switchPage.Show()
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
		inputPath := strings.TrimSpace(pathToVideo.Text)
		cmd := exec.Command("ls", inputPath)
		_, err := cmd.Output()
		if err != nil {
			isExits = false
			isFile = false
			isMP4File = false
		} else {
			isExits = true
			s, _ := os.Lstat(inputPath)
			isFile = !s.IsDir()
			isMP4File = path.Ext(path.Base(inputPath)) == ".mp4"
			if isFile && isMP4File {
				cmd := exec.Command("mplayer", "-identify", "-frames", "5", "-endpos", "0", "-vo", "null", inputPath)
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
			transferVideo(inputPath, &config, sshClient)
			label.SetText("Successfully transfer the video to the development board, you can click the button to open the browser")
			localVideoPage.Add(btnOpenBrowser)
			_, fileName := filepath.Split(inputPath)
			go execFaceRecognition(sshClient, &config, fileName, exitChan)
		}
	}

	localVideoPage.Add(form)

	mainWindow.Resize(fyne.NewSize(1000, 1000))
	mainWindow.ShowAndRun()
}
