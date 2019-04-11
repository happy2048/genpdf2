package main
import(
	"github.com/jessevdk/go-flags"
	"net/http"
	"fmt"
	"io"
	"bytes"
	"os"
	"log"
	"encoding/json"
	"io/ioutil"
)
type ReturnData struct {
	Msg string `json: "msg"`
	Data string `json: "data"`
	Code string `json: "code"`
}
type PdfClient struct {
	Content string 
	Type string
	Args string
}
type Options struct {
	Host string `short:"H" long:"host" description:"give the server ip which is running wkhtmltopdf." default:"127.0.0.1"`
	Port string `short:"P" long:"port" description:"give the server service listen port." default:"6660"`
	PdfArgs string `short:"a" long:"args" description:"give the pandoc args,eg: 'a::--toc -N' or 'c::--toc -N','a' \n is explained 'append','c' is explained 'change'." default:""`
	LatexTemp string `short:"t" long:"template" description:"give the template file for pandoc." default:""`
}
func main() {
	opt,args := NewOptions()
	opt.Check(args)
	PostGeneratePdfReq(args[0],opt.PdfArgs,opt.LatexTemp,opt.Host,opt.Port,args[1])
}
func PostGeneratePdfReq(con,args,temp,server,port,out string) {
	url := "http://" + server + ":"+ port + "/generate"
	data := make(map[string]string)
	tmpdata,err := ioutil.ReadFile(con)
	if err != nil {
		log.Printf("read %s failed,reason: %s\n",con,err.Error())
		return 
	}
	con = string(tmpdata)
	if temp != "" {
		tdata,err := ioutil.ReadFile(temp)
		if err != nil {
			log.Printf("read %s failed,reason: %s\n",temp,err.Error())
			return 
		}
		temp = string(tdata)
	}
	data["content"] = con
	data["args"] = args
	data["template"] = temp
	bytesData,err := json.Marshal(data)
	if err != nil {
		log.Printf("json marshal failed,reason: %s\n",err.Error())
		return
	}
	redata,err := Operate("POST",url,bytesData)
	if err != nil {
		log.Printf("http request failed,reason: %s\n",err.Error())
		return 
	}
	var parse ReturnData
	err = json.Unmarshal([]byte(redata),&parse)
	if err != nil {
		log.Printf("json unmarshal return data failed,reason: %s\n",err.Error())
		return
	}
	if parse.Code == "1000" {
		url = "http://" + server + ":"+ port + "/pdf/" + parse.Data
		res, err := http.Get(url)
		if err != nil {
			log.Printf("get pdf file failed,reason: %s\n",err.Error())
			return
		}
		f,err := os.Create(out)
		if err != nil {
			log.Printf("get pdf file failed,reason: %s\n",err.Error())
			return

		}
		io.Copy(f,res.Body)
	}else {
		log.Printf("get pdf file failed,reason: %s\n",parse.Msg)
	}
}

func GetOsEnv(env string) string {
    return os.Getenv(env)
}
func Operate(method,url string,data []byte) (string,error) {
    client := &http.Client{}
    var request *http.Request
    var err error
    if string(data) == "" {
        request,err = http.NewRequest(method,url,nil)
    }else {
        request,err = http.NewRequest(method,url,bytes.NewReader(data))
    }
    request.Header.Set("Connection", "keep-alive")
	if method == "POST" {
		request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
    response,err := client.Do(request)
    if err != nil {
        return "",err
    }
    if response.StatusCode == 200 {
        body,err := ioutil.ReadAll(response.Body)
        if err != nil {
         return "",err
        }
        return string(body),nil
    }
    return "",fmt.Errorf("%s","requst failure")
}
func NewOptions() (*Options,[]string) {
    var options Options
	pdata := flags.NewParser(&options, flags.Default)
	pdata.Usage = "[OPTIONS] INPUT [OUTPUT FILE]"
    args,err := pdata.Parse()
    if err != nil {
        if flagsErr, ok := err.(*flags.Error);ok && flagsErr.Type == flags.ErrHelp {
            os.Exit(0)
        }else {
			fmt.Println(err.Error())
        	os.Exit(1)
		}
    }
	if len(args) == 1 {
		args = append(args,"generate.pdf")
	}
    return &options,args
}


func (opt *Options) Check(args []string) {
    if len(os.Args) == 1 {
        fmt.Printf("Error: you should give some options,plese use -h or --help to get usage.\n")
        os.Exit(1)
    }
	if len(args) == 0 {
		fmt.Printf("Error: you should give the input resource (a html file or a url) and the name of output pdf file.\n")
		os.Exit(2)
	}

}

func CheckFileExist(filename string) bool {
    var exist = true
    _,err := os.Stat(filename)
    if os.IsNotExist(err) {
        exist = false
    }
    return exist
}

