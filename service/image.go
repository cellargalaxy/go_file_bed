package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"math"
)

func AddImageExtension(ctx context.Context, filePath string) string {
	return fmt.Sprintf("%s.%+v", filePath, config.Config.ImageSaveFormat)
}

func CompressionImage(ctx context.Context, buffer *bytes.Buffer) (*bytes.Buffer, error) {
	imageBytes := buffer.Bytes()
	img, err := imaging.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("压缩图片，图片解码异常")
		return nil, fmt.Errorf("压缩图片，图片解码异常: %+v", err)
	}

	imageSize := len(imageBytes)
	encodeOption := createJPEGQuality(ctx, imageSize)

	newBuffer := &bytes.Buffer{}
	err = imaging.Encode(newBuffer, img, config.Config.ImageSaveFormat, encodeOption)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("压缩图片，图片压缩异常")
		return nil, fmt.Errorf("压缩图片，图片压缩异常: %+v", err)
	}
	return newBuffer, nil
}

// jpeg的图片质量范围为[min,max]
// min的取值为：分别使用几M以及几十M的图片，从图片质量100开始往下调，直到调到最小的能容忍的值，为min，默认20
// max的取值为：分别使用几十K以及几百K的图片，从图片质量100开始往下调，直到调到最小的能容忍的值，为max，默认80
// quality=min+(max-min)*(0.99^(size/targetSize))
// 大小     10K          200K  500K         1M           5M           10M          30M          50M          100M
// ratio    0.999497609  0.99  0.975187187  0.949843809  0.773145054  0.597753274  0.21358261   0.076314984  0.005823977
// quality  79.96985657  79.4  78.51123123  76.99062854  66.38870321  55.86519643  32.81495663  24.57889903  20.34943861
func createJPEGQuality(ctx context.Context, size int) imaging.EncodeOption {
	power := float64(size) / config.Config.ImageTargetSize
	qualityRatio := math.Pow(0.99, power)
	quality := int(config.Config.JpegMinQuality + (config.Config.JpegMaxQuality-config.Config.JpegMinQuality)*qualityRatio)
	logrus.WithFields(logrus.Fields{"power": power, "qualityRatio": qualityRatio, "quality": quality}).Info("JPEG图片质量")
	return imaging.JPEGQuality(quality)
}
