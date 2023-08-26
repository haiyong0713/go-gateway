package generator

import (
	"context"
	"fmt"
)

type UdfTaskGenerator struct {
	TmplFilePath string
	GoFilePath   string
	ProjectName  string
	ServiceName  string
	ServicePath  string
}

func NewUdfTaskGenerator(ctx context.Context, codeInfo *CodeInfo) (taskGenerator *UdfTaskGenerator, err error) {
	return &UdfTaskGenerator{
		TmplFilePath: codeInfo.TmplFilePath,
		GoFilePath:   codeInfo.GoFilePath,
		ProjectName:  codeInfo.ProjectName,
		ServiceName:  codeInfo.ServiceName,
		ServicePath:  codeInfo.ServicePath,
	}, nil
}

func (p *UdfTaskGenerator) Generate(ctx context.Context) (err error) {
	fmt.Println("Generate UdfTask")

	return
}

func (p *UdfTaskGenerator) copyFile(src, srcFile, dest, destFile string) (err error) {
	err = CopyFile(src, srcFile, dest, destFile)
	return
}
