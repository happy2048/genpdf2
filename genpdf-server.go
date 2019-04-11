package main
import(
	"os/exec"
	"math/rand"
	"path"
	"time"
	"bytes"
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
	"io"
	"os"
	"log"
	"io/ioutil"
)
type HtmlInfo struct {
	Type string `json: "type"`
	Content string `json: "content"`
	Args string `json: "args"`
	Template string `json: "template"`
	Name string `json: "name"`
}
type FileInfo struct {
	FileName string `json:"filename"`
	Content  string `json: "content"`
}
type ReturnData struct {
	Msg string `json: "msg"`
	Data string `json: "data"`
    Code  string `json: "code"`
}
func main() {
	if GetOsEnv("PANDOC_CMD") == "" {
		os.Setenv("PANDOC_CMD","pandoc --pdf-engine=xelatex")
	}
	if GetOsEnv("PANDOC_DEFAULT_ARGS") == "" {
		os.Setenv("PANDOC_DEFAULT_ARGS","-V fontsize=12pt  --listings ")
	}
	if GetOsEnv("TMP_PATH") == "" {
		os.Setenv("TMP_PATH","/tmp/pdf")
	}
	if GetOsEnv("PORT") == "" {
		os.Setenv("PORT","6660")
	}
	if  state,_ := PathExists(GetOsEnv("TMP_PATH"));state == false {
		os.Mkdir(GetOsEnv("TMP_PATH"),0644)
	}
	http.HandleFunc("/generate",HandleHtml)
	http.HandleFunc("/deletefiles",HandleDeleteTmpFiles)
	http.Handle("/pdf/", http.StripPrefix("/pdf/", http.FileServer(http.Dir(GetOsEnv("TMP_PATH") + "/"))))
	log.Printf("service is starting")
	err := http.ListenAndServe(":" + GetOsEnv("PORT"),nil)
	if err != nil {
		log.Printf("Error: ",err.Error())
		return 
	}
}
		
func HandleDeleteTmpFiles(w http.ResponseWriter,req *http.Request) {
	if req.Method == "GET" {
		if req.Host != "localhost:" + GetOsEnv("PORT") && req.Host != "127.0.0.1:" + GetOsEnv("PORT") {
			ReturnValue(w,"1100","","you have not privilege to execute this option.")
			return
			
		}
		_,_,err := RunCmd("rm -rf " + GetOsEnv("TMP_PATH") + "/*")
		if err != nil {
			ReturnValue(w,"1100","",err.Error())
			return
		}
		ReturnValue(w,"1000","","delete tmp files succeed.")
		return
	}else {
		ReturnValue(w,"1100","","invalid method")
		return
	}
}
func HandleHtml(w http.ResponseWriter,req *http.Request)  {
	if req.Method == "POST" {
		body,_ := ioutil.ReadAll(req.Body)
		var htmlInfo HtmlInfo
		if err := json.Unmarshal(body,&htmlInfo);err == nil {
			pdf,err := CreatePdf(htmlInfo.Content,htmlInfo.Args,htmlInfo.Template)
			if err != nil {
				ReturnValue(w,"1100","",err.Error())
				return
			}
			ReturnValue(w,"1000",pdf,"")
		}else {
			ReturnValue(w,"1100","",err.Error())
			return 
		}
	}else {
		ReturnValue(w,"1100","","http method is invalid,you should use POST.")
		return
	}
}
func CreatePdf(content,args,template string) (string,error) {
	fid := GetRandomString(10)
	md := fid + ".md"
	pdf := fid + ".pdf"
	temp := fid + ".tex"
	tmpPath := GetOsEnv("TMP_PATH")
	if tmpPath == "" {
		return "",fmt.Errorf("%s","tmp path is not found.")
	}
	err := ioutil.WriteFile(path.Join(GetOsEnv("TMP_PATH"),md),[]byte(content),os.ModeAppend)
	if  err != nil {
		return  "",err
	}
	if template != "" {
		err := ioutil.WriteFile(path.Join(GetOsEnv("TMP_PATH"),temp),[]byte(template),os.ModeAppend)
		if err != nil {
			return "",err
		}
	}
	cmdStr := GetOsEnv("PANDOC_CMD")
	if args != "" {
		argsList := strings.Split(args,"::")
		if len(argsList) == 2 {
			if argsList[0] == "a" {
				cmdStr +=  " " + argsList[1] + " " +  GetOsEnv("PANDOC_DEFAULT_ARGS")
			}else if   argsList[0] == "c" {
				cmdStr +=  " " + argsList[1]
			}else {
				log.Printf("warning: the format of args is invalid.\n")
				cmdStr +=  " "  + GetOsEnv("PANDOC_DEFAULT_ARGS") 
			}
		}else {
			return  "",fmt.Errorf("invalid format of pandoc options,you should give them like 'a:: --toc -N' or 'c:: --toc -N'")
		}
	}else {
		cmdStr += " " + GetOsEnv("PANDOC_DEFAULT_ARGS")
	}
	if template != "" {
		cmdStr += " --template=" + path.Join(GetOsEnv("TMP_PATH"),temp)
	}else {
		cmdStr += " --template=/root/template.tex"
	}
	cmdStr += " -o " + path.Join(GetOsEnv("TMP_PATH"),pdf) +  " "  +   path.Join(GetOsEnv("TMP_PATH"),md)
	log.Printf("current cmd: %s\n",cmdStr)
	out,stderr,err := RunCmd(cmdStr)
	log.Printf("\n%s\n",stderr)
	log.Printf("\n%s\n",out)
	if err != nil {
		return out,err
	}
	return pdf,err
}
func GetOsEnv(env string) string {
    return os.Getenv(env)
}
func ReturnValue(w http.ResponseWriter,code string,data string,msg string) {
    redata,err := json.Marshal(&ReturnData{Code: code,Msg: msg,Data: data})
    if err != nil {
		log.Printf("%s",err.Error())
        return
    }
    io.WriteString(w,string(redata))
}
func RunCmd(cmdStr string) (string,string,error) {
	cmd := exec.Command("/bin/sh","-c",cmdStr)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	return out.String(),errOut.String(),err
}

func GetRandomString(length int64) string{
   str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
   bytes := []byte(str)
   result := []byte{}
   r := rand.New(rand.NewSource(time.Now().UnixNano()))
   for i := int64(0); i < length; i++ {
      result = append(result, bytes[r.Intn(len(bytes))])
   }
   return string(result)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

