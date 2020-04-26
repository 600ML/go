package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type point struct {
	X int
	Y int
}

var adb = "/opt/genymotion/tools/adb"

func main() {
	devList := deviceList()
	fmt.Println(devList)
	if len(devList) == 0 {
		fmt.Println("no device found")
		return
	}

	if len(devList) > 1 {
		fmt.Println("multiple device found")
		return
	}

	fmt.Println(connect(devList[0]))
	fmt.Println(execute(adb, "root"))

	//love := point{X: 1300, Y: 1200}
	//fmt.Println(tap(love))

	for {
		commentFlag := false
		comment := point{X: 1300, Y: 1400}
		fmt.Println(tap(comment))

		time.Sleep(time.Second * 1)

		for {
			commentFileName := screenShot()

			commentFile, err := os.Open(commentFileName)
			if err != nil {
				fmt.Printf("read file(%s) failed, err:%s", commentFileName, err.Error())
				return
			}

			commentImg, err := png.Decode(commentFile)
			if err != nil {
				commentFile.Close()
				fmt.Printf("png.decoe file(%s) failed, err:%s", commentFileName, err.Error())
				return
			}

			// 截图实际像素 => 模拟器实际像素 1440x2560
			// 1358 => 516 比例2.63
			// 715 => 276 比例 2.59
			commentX := 112
			for commentY := 675; commentY < commentImg.Bounds().Max.Y-183; commentY++ {
				commentRed, commentGreen, commentBlue := getRGB(commentImg, commentX, commentY)
				if commentRed == 255 && commentGreen == 255 && commentBlue == 255 {
					continue
				}

				commentFlag = false
				tap(point{commentX, commentY + 15})
				time.Sleep(time.Second * 1)

				homeFileName := screenShot()

				homeFile, err := os.Open(homeFileName)
				if err != nil {
					fmt.Printf("read file(%s) failed, err:%s", homeFileName, err.Error())
					return
				}

				homeImg, err := png.Decode(homeFile)
				if err != nil {
					homeFile.Close()
					fmt.Printf("png.decoe file(%s) failed, err:%s", homeFileName, err.Error())
					return
				}

				followX := 730
				followY := 656
				red, green, blue := getRGB(homeImg, followX, followY) //找到关注按钮
				if red == 254 && green == 44 && blue == 85 {
					tap(point{followX, followY}) //点击关注

					backX := 85
					backY := 178
					tap(point{backX, backY}) //点击返回
					time.Sleep(time.Second)
				}

				os.Remove(homeFileName)
				commentY += 130
			}
			commentFile.Close()

			os.Remove(commentFileName)
			if commentFlag {
				break
			}
			swipe(100, 1730, 100, 715)
			time.Sleep(time.Second * 3)
		}

		swipe(100, 1730, 100, 715)
		time.Sleep(time.Second * 3)
	}
}

func deviceList() []string {
	result := execute(adb, "devices")

	reg, err := regexp.Compile(`([\d:\.]+)`)
	if err != nil {
		return []string{err.Error()}
	}

	return reg.FindAllString(result, -1)
}

func connect(str string) string {
	return execute(adb, "connect", str)
}

func screenShot() string {
	fileName := fmt.Sprintf("%d.png", time.Now().UnixNano())
	sdcardPath := fmt.Sprintf("/sdcard/screen_shot/%s", fileName)
	result := execute(adb, "shell", "screencap", "-p", sdcardPath)
	fmt.Println("screencap, result:", result)

	result = execute(adb, "pull", sdcardPath)
	fmt.Println("pull image to local, result:", result)

	execute(adb, "shell", "rm", sdcardPath)

	dir := "screen_shot"
	_, err := os.Stat(dir)
	if err != nil {
		err := os.Mkdir(dir, os.FileMode(0777))
		if err != nil {
			fmt.Printf("mkdir dir failed, dir:%s, err:%s", dir, err.Error())
		}
	}

	localFileName := dir + "/" + fileName
	execute("mv", fileName, localFileName)
	return localFileName
}

func tap(p point) string {
	args := []string{
		"shell",
		"input",
		"tap",
		fmt.Sprintf("%d", p.X),
		fmt.Sprintf("%d", p.Y),
	}
	return execute(adb, args...)
}

func execute(cmdLine string, args ...string) string {
	cmd := exec.Command(cmdLine, args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return err.Error()
	}

	if err := cmd.Wait(); err != nil {
		return err.Error()
	}

	return out.String()
}

func getRGB(img image.Image, x, y int) (uint8, uint8, uint8) {
	theColor := img.At(x, y)
	return theColor.(color.NRGBA).R, theColor.(color.NRGBA).G, theColor.(color.NRGBA).B
}

func swipe(x1, y1, x2, y2 int) string {
	args := []string{
		"shell",
		"input",
		"swipe",
		fmt.Sprintf("%d", x1),
		fmt.Sprintf("%d", y1),
		fmt.Sprintf("%d", x2),
		fmt.Sprintf("%d", y2),
	}
	return execute(adb, args...)
}
