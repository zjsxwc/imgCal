package main

import (
	"flag"
	"fmt"
"github.com/disintegration/imaging"
	"image/color"
	"io"
	"net/http"
	"strconv"
)

func calculateProportion(ioReader io.Reader) {
	// 读取图片
	src, err := imaging.Decode(ioReader)
	if err!= nil {
		fmt.Println("Error opening image:", err)
		return
	}

	// 转换为灰度图
	grayImg := imaging.Grayscale(src)

	// 计算直方图
	histogram := make([]int, 256)
	for y := 0; y < grayImg.Bounds().Dy(); y++ {
		for x := 0; x < grayImg.Bounds().Dx(); x++ {
			pixel := grayImg.At(x, y)
			c := color.NRGBAModel.Convert(pixel).(color.NRGBA)
			y,_,_ := color.RGBToYCbCr(c.R, c.G, c.B)
			grayValue := y
			histogram[grayValue]++
		}
	}

	// 找到最大数量的灰度级
	maxbin := 0
	maxCount := 0
	for i, count := range histogram {
		if count > maxCount {
			maxCount = count
			maxbin = i
		}
	}

	lowerBound := maxbin - 30
	upperBound := maxbin + 30

	if lowerBound < 0 {
		lowerBound = 0
	}
	if upperBound > 255 {
		upperBound = 255
	}

	// 计算指定范围内的像素数量
	pixelCountInRange := 0
	for i := lowerBound; i <= upperBound; i++ {
		pixelCountInRange += histogram[i]
	}

	// 计算图片总像素数量
	totalPixels := grayImg.Bounds().Dx() * grayImg.Bounds().Dy()

	// 计算比重
	proportion := float64(pixelCountInRange) / float64(totalPixels)
	p, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", proportion), 64)

	//fmt.Printf("从 maxbin - 35 到 maxbin + 35 之间的像素数量占图片总像素点的比重为: %.2f\n", p)
	if p >= 0.95 {
		fmt.Printf("SolidColorCloth");
	} else {
		fmt.Printf("PatternedCloth");
	}
}

//GO111MODULE=off go get github.com/disintegration/imaging
//GO111MODULE=off go run imgCal/imgCal.go -imgUrl https://xxx/yyy.jpg
//GO111MODULE=off go build imgCal.go
//GO111MODULE=off go build -gcflags="-N -l" -ldflags "-s -w" imgCal.go
func main() {
	var imgUrl string;
	flag.StringVar(&imgUrl, "imgUrl", "", "image url");
	flag.Parse();
	if len(imgUrl) <= 0 {
		return;
	}
	response, err := http.Get(imgUrl);
	if err!= nil {
		return;
	}
	defer response.Body.Close();
	// 检查响应状态码
	if response.StatusCode!= http.StatusOK {
		return
	}
	ioReader := response.Body.(io.Reader);

	calculateProportion(ioReader)  // 替换为实际的图片路径
}

