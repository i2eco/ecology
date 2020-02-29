package captcha

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"github.com/i2eco/ecology/appgo/pkg/code"
	"github.com/i2eco/ecology/appgo/router/core"
)

type CaptchaResponse struct {
	CaptchaId string `json:"captchaId"`
	ImageUrl  string `json:"imageUrl"`
}

func Captcha(c *core.Context) {
	captchaId := captcha.New()

	if captchaId == "" {
		c.JSONErr(code.MsgErr, nil)
		return
	}
	c.JSONOK(gin.H{
		"captchaId": captchaId,
		"image":     "/captcha/" + captchaId + ".png",
	})
}

//func VerifyCaptcha(c *gin.Context) {
//	baseResponse := model.NewBaseResponse()
//	captchaId := context.Request.FormValue("captchaId")
//	value := context.Request.FormValue("value")
//	if captchaId == "" || value == "" {
//		baseResponse.GetFailureResponse(model.QUERY_PARAM_ERROR)
//	} else {
//		if captcha.VerifyString(captchaId, value) {
//			baseResponse.GetSuccessResponse()
//			baseResponse.Message = "验证成功"
//		} else {
//			baseResponse.GetFailureResponse(model.CAPTCHA_ERROR)
//		}
//	}
//	c.JSON(baseResponse)
//}

func CaptchaPng(c *gin.Context) {
	dir, file := path.Split(c.Request.URL.Path)
	ext := path.Ext(file)
	id := file[:len(file)-len(ext)]
	if ext == "" || id == "" {
		http.NotFound(c.Writer, c.Request)
		return
	}
	if c.Query("reload") != "" {
		captcha.Reload(id)
	}
	lang := strings.ToLower(c.Query("lang"))
	download := path.Base(dir) == "download"

	fmt.Println("id------>", id)

	if Serve(c.Writer, c.Request, id, ext, lang, download, captcha.StdWidth, captcha.StdHeight) == captcha.ErrNotFound {
		http.NotFound(c.Writer, c.Request)
	}
}

func Serve(w http.ResponseWriter, r *http.Request, id, ext, lang string, download bool, width, height int) error {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	var content bytes.Buffer
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
		err := captcha.WriteImage(&content, id, width, height)
		if err != nil {
			fmt.Println("err", err)
		}
	case ".wav":
		w.Header().Set("Content-Type", "audio/x-wav")
		captcha.WriteAudio(&content, id, lang)
	default:
		return captcha.ErrNotFound
	}

	if download {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	http.ServeContent(w, r, id+ext, time.Time{}, bytes.NewReader(content.Bytes()))
	return nil
}
