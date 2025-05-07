package mint

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
)

func MintNFT(ctx context.Context, did string) ([]byte, error) {
	imageNum := rand.Intn(10) + 1
	imagePath := filepath.Join("pkg/mint/fake", fmt.Sprintf("%d.png", imageNum))

	// 读取图片文件
	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("读取图片失败: %w", err)
	}

	return imageData, nil
}
