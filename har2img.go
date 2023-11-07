package main

import (
	"log"
	"strings"
	"encoding/json"
	"encoding/base64"
	"os"
	"fmt"
)

/**
* 保存har文件中的所有图片，使用方式：
* 1. 无参数使用命令：go run har2img.go，为当前文件夹下的每个har文件创建文件夹，将har文件中的所有jpg格式的图片存到对应的文件夹
* 2. 有参数使用命令：go run har2img.go srchar destDir imgFormat，为srchar中的所有imgFormat格式的图片存到destDir中去
*/

type Content struct {
	MimeType string `json:mimeType`
	Text string `json:text`
}

type Response struct {
	Content Content `json:content`
}

type Entry struct {
	Pageref string `json:pageref`
	ResourceType string `json:_resourceType`
	Response Response `json:response`
}

type Har struct {
	Log *HarLog `json:log`
}

type HarLog struct {
	Entries []*Entry `json:entries`
}

func main() {
    	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]
	if len(args) > 0 && len(args) != 3 {
		panic("invalid program arguments")
	}

	for _, file := range files {
		fn := file.Name()
		if !strings.HasSuffix(fn, ".har") {
			continue
		}

		if len(args) > 0 && fn != args[0] {
			continue
		}

		f, err := os.Open(fn)
		if err != nil {
			log.Fatal(err)
		}

		fi, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}

		fsize := fi.Size()
		bs := make([]byte, fsize)
		_, err = f.Read(bs)
		if err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}

		har := &Har{}
		if err := json.Unmarshal(bs, har); err != nil {
			panic(err)
		}

		dir := strings.Replace(fn, ".har", "", 1)
		if len(args) > 1 {
			dir = args[1]
		}
		if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
			panic(err)
		}

		mimeType := "image/jpeg"
		suffix := "jpg"
		if len(args) > 2 {
			suffix = args[2]
			mimeType = map[string]string{
				"jpg": "image/jpeg",
				"png": "image/png",
			}[suffix]

			if mimeType == "" {
				panic("only support jpg and png now")
			}
		}

		counter := 0
		for _, entry := range har.Log.Entries {
			if entry.Response.Content.MimeType != mimeType {
				continue
			}
			fmt.Println(entry.Response.Content.MimeType)

			counter++
			f, err = os.OpenFile(fmt.Sprintf("%s/%d.%s", dir, counter, suffix), os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				panic(err)
			}


			decoded, err := base64.StdEncoding.DecodeString(entry.Response.Content.Text)
			if err != nil {
				panic(err)
			}
			if _, err := f.Write(decoded); err != nil {
				panic(err)
			}

			if err := f.Close(); err != nil {
				panic(err)
			}
		}
	}
}
