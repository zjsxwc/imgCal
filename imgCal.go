package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/disintegration/imaging"
	"math/cmplx"
)



func estimate(ioReader io.Reader)  {
	// 读取图片
	src, err := imaging.Decode(ioReader)
	if err!= nil {
		fmt.Println("Error opening image:", err)
		return
	}

	histogram, grayImg := calculateHistogram(src)

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
		fmt.Printf("不是花型 纯色图");
	} else {
		splitAndCalculateHistograms(src);
	}
}

func calculateHistogram(src image.Image) ([]int, *image.NRGBA) {
	// 转换为灰度图
	grayImg := imaging.Grayscale(src);
	// 计算直方图
	histogram := make([]int, 256);
	for y := 0; y < grayImg.Bounds().Dy(); y++ {
		for x := 0; x < grayImg.Bounds().Dx(); x++ {
			pixel := grayImg.At(x, y);
			c := color.NRGBAModel.Convert(pixel).(color.NRGBA);
			y,_,_ := color.RGBToYCbCr(c.R, c.G, c.B);
			grayValue := y;
			histogram[grayValue]++;
		}
	}
	return histogram, grayImg;
}


func splitAndCalculateHistograms(img image.Image) {
    width := img.Bounds().Dx()
    halfWidth := width / 2

    leftHalf := img.(interface {
        SubImage(r image.Rectangle) image.Image
    }).SubImage(image.Rect(0, 0, halfWidth, img.Bounds().Dy()))
    rightHalf := img.(interface {
        SubImage(r image.Rectangle) image.Image
    }).SubImage(image.Rect(halfWidth, 0, width, img.Bounds().Dy()))

    leftHistogram,_ := calculateHistogram(leftHalf)
    rightHistogram,_ := calculateHistogram(rightHalf)

   
	lmaxbin := 0
	lmaxCount := 0
	for i, count := range leftHistogram {
		if count > lmaxCount {
			lmaxCount = count
			lmaxbin = i
		}
	}
	
	rmaxbin := 0
	rmaxCount := 0
	for i, count := range rightHistogram {
		if count > rmaxCount {
			rmaxCount = count
			rmaxbin = i
		}
	}
	

	if math.Abs(float64(rmaxbin) - float64(lmaxbin)) > 100 {
		fmt.Printf("不是花型 左右直方图最大灰度本身偏差大于100")
	} else {
		p := float64(rmaxCount)/float64(lmaxCount)
		if (p >= 0.43) && (p <= 2.28) {
			//fmt.Printf("左右直方图数组 傅里叶变换比较\n")

			// 生成两个包含 256 个元素的数组
			array1 := make([]complex128, 256)
			array2 := make([]complex128, 256)
			for i := 0; i < 256; i++ {
				array1[i] = complex(float64(leftHistogram[i]), 0)
				array2[i] = complex(float64(rightHistogram[i]), 0)
			}

			// 进行 FFT
			fft1 := fft(array1)
			fft2 := fft(array2)

			// 计算相似性
			similarity := correlation(fft1, fft2)
			if ((similarity < 1.12) && (similarity > 0.92) ) {
				fmt.Printf("是花型")
			} else {
				fmt.Printf("不是花型 左右直方图数组 傅里叶变换比较 相似度大于1.12 或 小于0.92  相似性：%.2f", similarity)
			}


		} else {
			fmt.Printf("不是花型 左右直方图最大灰度的值偏差大于2.28倍")
		}
	}
	

}

// 计算两个复数数组的相关性
func correlation(a, b []complex128) float64 {
    var sumAB, sumA2, sumB2 complex128
    for i := 0; i < len(a); i++ {
        sumAB += a[i] * b[i]
        sumA2 += a[i] * a[i]
        sumB2 += b[i] * b[i]
    }
    magnitudeAB := cmplx.Abs(sumAB)
    magnitudeA :=  cmplx.Abs(cmplx.Sqrt(sumA2))
    magnitudeB := cmplx.Abs(cmplx.Sqrt(sumB2))
    return (magnitudeAB) / (magnitudeA * magnitudeB)
}

// 快速傅里叶变换函数
func fft(x []complex128) []complex128 {
    n := len(x)
    if n == 1 {
        return x
    }
    even := make([]complex128, n/2)
    odd := make([]complex128, n/2)
    for i := 0; i < n/2; i++ {
        even[i] = x[2*i]
        odd[i] = x[2*i+1]
    }
    fftEven := fft(even)
    fftOdd := fft(odd)
    omega := complex(1, 0)
    omegaN := cmplx.Exp(-2 * math.Pi * complex(0, 1) / complex(float64(n), 0))
    result := make([]complex128, n)
    for k := 0; k < n/2; k++ {
        t := omega * fftOdd[k]
        result[k] = fftEven[k] + t
        result[k+n/2] = fftEven[k] - t
        omega *= omegaN
    }
    return result
}

//go run imgCal.go -imgUrl https://xxx/xxx3f774.jpeg_thumbnail400.jpg
//go build imgCal.go
//go build -gcflags="-N -l" -ldflags "-s -w" imgCal.go
//  GOOS=linux GOARCH=amd64  go build -ldflags '-linkmode "external" -extldflags "-static -L/path/to/glibc2.19/lib"' main.go
//因为我们docker所以我们不用依赖cgo,无cgo的编译方式如下，推荐：
//GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o imgCal-nocgo -gcflags="-N -l" -ldflags "-s -w" imgCal.go
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

	estimate(ioReader)
}

