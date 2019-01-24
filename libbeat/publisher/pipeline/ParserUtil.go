package pipeline

import (
	"fmt"
	"time"
	"strings"
	"runtime"
	"strconv"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

const(
	CONVERSION_CHARS = "cCdFlLmMnprtxX"

	f_windows ="windows"
	f_message  = "message"
	f_msg="msg"//系统字段
	f_datetime = "datetime"//系统字段
	f_logdate="logdate"
	f_loglevel="loglevel"
	f_source  = "source"
	f_logformat="logformat"
	f_adflogformat="adflog"
	f_logpattern="logpattern"

)

type Log4jParser struct{
	event *beat.Event
	pattern string
	tokens []string
	filename string
	fieldsvalue common.MapStr

}

func (p *Log4jParser) SetPattern(pattern string) {
	p.pattern=pattern
	p.tokens=strings.Split(pattern,"%")
	p.fieldsvalue= common.MapStr{};

}

func (p *Log4jParser) SetFilename(name string) {
	p.filename=name
}

func  isConversionCharacter(res rune) bool {
	//.indexOf(res)
	return (strings.IndexRune(CONVERSION_CHARS,res) != -1);
}

func (p *Log4jParser) ParseMessage(msg string) bool{
	i:=0
	strmsg:=msg
	for i<len(p.tokens){
		if len(p.tokens[i])>0{
			if isConversionCharacter([]rune(p.tokens[i])[0]){
				strmsg = p.identifyConversionCharacter(p.tokens[i][0],strmsg,p.tokens[i])
			}
		}
		i=i+1
	}
	return true
}

func (p *Log4jParser)identifyConversionCharacter(c byte, msg string ,token string) string {
	switch c {
	case 'c':
		category,nmsg :=p.getProperty(msg, token, 'c')
		p.event.Fields.Put("CATEGORY",category)
		return  nmsg

	case 'C':
		class ,nmsg :=p.getProperty(msg, token, 'C')
		p.event.Fields.Put("CLASS",class)
		return  nmsg
	case 'd':
		datetime,nmsg,dtint,logdate:=p.getDate(msg, token)//, 'd')
		p.event.Fields.Put("STRDATETIME",datetime)
		p.event.Fields.Put(f_datetime,dtint)
		p.event.Fields.Put(f_logdate,logdate)
		fmt.Println("###logdate :"+logdate)
		return  nmsg
	case 'F':
		fdvalue,nmsg  :=p.getProperty(msg, token, 'F')
		p.event.Fields.Put("FILE",fdvalue)
		return  nmsg
	case 'l':
		fdvalue,nmsg  :=p.getProperty(msg, token, 'l')
		p.event.Fields.Put("LOCATION",fdvalue)
		return  nmsg
	case 'L':
		fdvalue,nmsg  :=p.getProperty(msg, token, 'L')
		p.event.Fields.Put("LINE",fdvalue)
		return  nmsg
	case 'm':
		fdvalue,nmsg  :=p.getProperty(msg, token, 'm')

		p.event.Fields.Put(f_msg,fdvalue)
		return  nmsg
	case 'M':
		fdvalue,nmsg  :=p.getProperty(msg, token, 'M')
		p.event.Fields.Put("METHOD",fdvalue)
		return  nmsg
	case 'n':
		p.event.Fields.Put("LINE_SEPERATOR","\n")
		return ""
	case 'p':
		fdvalue,nmsg  :=p.getProperty(msg, token, 'p')
		p.event.Fields.Put(f_loglevel,fdvalue)
		return  nmsg
	case 'r':
		fdvalue ,nmsg :=p.getProperty(msg, token, 'r')
		p.event.Fields.Put("ELAPSED",fdvalue)
		return  nmsg
	case 't':
		fdvalue,nmsg  :=p.getProperty(msg, token, 't')
		p.event.Fields.Put("THREAD",fdvalue)
		return  nmsg
	case 'x':
		fdvalue,nmsg  :=p.getProperty(msg, token, 'x')
		p.event.Fields.Put("NDC",fdvalue)
		return  nmsg
	case 'X':
		fdvalue,nmsg  :=p.getProperty(msg, token, 'X')
		p.event.Fields.Put("MDC",fdvalue)
		return  nmsg
	default:
		return ""
	}
	return ""
}
//return fieldvalue[datetime], substring,unixtime,logdate[es index]
func (p *Log4jParser) getDate(str string, token string)(string,string,int64,string) {
	dateformat:="yyyy-MM-dd HH:mm:ss,SSS"
	sep:=p.getSeparator(token);
	ind :=strings.Index(token, "{")
	if ind!=-1{
		spattern :=substring(token,ind + 1,strings.Index(token, "}"))
		if spattern=="ABSOLUTE"{
			dateformat="HH:mm:ss,SSS"
			datevalue:=substring(str,0,len(dateformat))
			msg:=substring(str,len(dateformat)+len(sep),len(str))

			s:=time.Now().Format("2006.01.02")
			fmt.Println(s)
			//e.Fields.Put(f_logdate,s)
			return datevalue,msg,0,s
		}else if spattern=="DATE"{
			dateformat = "dd MMM yyyy HH:mm:ss,SSS"
			datevalue:=substring(str,0,len(dateformat))
			msg:=substring(str,len(dateformat)+len(sep),len(str))

			dvaluei,s:=parse2unix(dateformat,datevalue)
			fmt.Println(s)
			return datevalue,msg,dvaluei,s
		}else if spattern=="ISO8601"{
			datevalue:=substring(str,0,len(dateformat))
			msg:=substring(str,len(dateformat)+len(sep),len(str))
			dvaluei,s:=parse2unix(dateformat,datevalue)
			return datevalue,msg,dvaluei,s
		}else{
			dateformat = spattern
			datevalue:=substring(str,0,len(dateformat))
			msg:=substring(str,len(dateformat)+len(sep),len(str))
			dvaluei,s:=parse2unix(dateformat,datevalue)
			return datevalue,msg,dvaluei,s
		}
	}
 	fdvalue,msg:=p.getProperty(str,token,'d')
	return fdvalue,msg,0,time.Now().Format("yyyy.MM.dd")
}

func parse2unix(pattern string,dvalue string) (int64,string){
	if checkPattern(pattern){
		ystr:=substring(dvalue,strings.Index(pattern,"yyyy"),strings.Index(pattern,"yyyy")+4)
		yeari,_:=strconv.Atoi(ystr)
		dstr:=substring(dvalue,strings.Index(pattern,"dd"),strings.Index(pattern,"dd")+2)
		di,_:=strconv.Atoi(dstr)
		hstr:=substring(dvalue,strings.Index(pattern,"HH"),strings.Index(pattern,"HH")+2)
		hi,_:=strconv.Atoi(hstr)
		minstr:=substring(dvalue,strings.Index(pattern,"mm"),strings.Index(pattern,"mm")+2)
		mini,_:=strconv.Atoi(minstr)
		sstr:=substring(dvalue,strings.Index(pattern,"ss"),strings.Index(pattern,"ss")+2)
		si,_:=strconv.Atoi(sstr)

		mstr:="0"
		mi:=0
		if strings.Contains(pattern,"MMM"){
			mstr=substring(dvalue,strings.Index(pattern,"MMM"),strings.Index(pattern,"MMM")+3)
			mi=convertMMM2Int(mstr)
			if mi==0{
				return 0,time.Now().Format("2006.01.02")
			}
		}else if strings.Contains(pattern,"MM"){
			mstr=substring(dvalue,strings.Index(pattern,"MM"),strings.Index(pattern,"MM")+2)
			mi,_ =strconv.Atoi(mstr)
		}
		t := time.Date(yeari, time.Month(mi), di, hi, mini, si, 0, time.Local)
		if strings.Contains(pattern,"SSS"){
			milsecstr:=substring(dvalue,strings.Index(pattern,"SSS"),strings.Index(pattern,"SSS")+3)
			milseci,_:=strconv.Atoi(milsecstr)
			return t.Unix()*1000+int64(milseci),t.Format("2006.01.02")
		} else {
			return t.Unix()*1000,t.Format("2006.01.02")
		}
	}else {
		return 0,time.Now().Format("2006.01.02")
	}
}

func checkPattern(pattern string) bool{
	if strings.Contains(pattern,"yyyy") && strings.Contains(pattern,"dd") && strings.Contains(pattern,"HH") &&
		strings.Contains(pattern,"mm") && strings.Contains(pattern,"ss") {
		if  strings.Contains(pattern,"MMM") || strings.Contains(pattern,"MM"){
			return true
		}
		return false
	}
	return false
}

func convertMMM2Int(month string) int{
	if strings.Contains(month,"Jan") ||  strings.Contains(month,"一月") {
		return 1
	} else if strings.Contains(month,"Feb")||  strings.Contains(month,"二月") {
		return 2
	}else if strings.Contains(month,"Mar")||  strings.Contains(month,"三月") {
		return 3
	}else if strings.Contains(month,"Apr")||  strings.Contains(month,"四月") {
		return 4
	}else if strings.Contains(month,"May")||  strings.Contains(month,"五月") {
		return 5
	}else if strings.Contains(month,"Jun")||  strings.Contains(month,"六月") {
		return 6
	}else if strings.Contains(month,"Jul")||  strings.Contains(month,"七月") {
		return 7
	}else if strings.Contains(month,"Aug")||  strings.Contains(month,"八月") {
		return 8
	}else if strings.Contains(month,"Sep")||  strings.Contains(month,"九月") {
		return 9
	}else if strings.Contains(month,"Oct")||  strings.Contains(month,"十月") {
		return 10
	}else if strings.Contains(month,"Nov")||  strings.Contains(month,"十一月") {
		return 11
	}else if strings.Contains(month,"Dec")||  strings.Contains(month,"十二月") {
	return 12
	}
	return 0
}

func (p *Log4jParser) getProperty(str string,token string,separator byte)( string,string ){
	sep:=p.getSeparator(token);

	if len(sep)<1{
		//不继续解析
		return str, ""
	}

	if strings.Contains(str,sep)	{
		value:=substring(str,0,strings.Index(str,sep))
		msg:=substring(str,strings.Index(str,sep)+len(sep),len(str))
		return value,msg
	}else  {
		fmt.Println("###getProperty not contain sep")
		return str, ""
	}
}
func (p *Log4jParser) getSeparator(token string) string{
	if len(token)<2{
		return ""
	}
	ind :=strings.Index(token, "}")
	if ind!=-1{
		if ind== len(token)-1{
			return "";
		}
		return substring(token,ind+1,len(token))
	}
	return substring(token,1,len(token))

}
//rune

func ParselogMsg(e *beat.Event) error{

	logformat,err:=e.Fields.GetValue(f_logformat)
	if err!=nil{
		return nil
	}
	logformatstr:=logformat.(string)
	if logformatstr==f_adflogformat{
		return ParseAdfMsg(e)
	}
	ok, _ :=e.Fields.HasKey(f_logpattern)
	if ok{
		return Parselog4jMsg(e)
	}

	return nil
}

func Parselog4jMsg(e *beat.Event) error {
	defer  func(){
		r:=recover()
		if r!=nil{
			fmt.Println("internal error: %v", r)
			fmt.Errorf("internal error: %v", r)
		}
	}()

	logpattern,err := e.Fields.GetValue(f_logpattern)
	if err!=nil{
		return nil
	}
	logpatterstr:=logpattern.(string)

	source,err:=e.Fields.GetValue(f_source)
	if err!=nil{
		return fmt.Errorf("not existed field source")
	}
	p:= &Log4jParser{
		event:e,
	}
	p.SetFilename(source.(string))
	p.SetPattern(logpatterstr)

	msg,err:=e.Fields.GetValue(f_message)
	p.ParseMessage(msg.(string))
//%d{ISO8601} %p [%t] %c{2}: %m%n

	return nil
}

func ParseAdfMsg(e *beat.Event) error {
	defer func(){
		r := recover()
		if r!=nil{
			fmt.Errorf("parse adf log error")
		}

	}()
	msg,err:=e.Fields.GetValue(f_message)
	if err!=nil{
		return fmt.Errorf("not existed field message")
	}
	if len(msg.(string))<13 {
		return fmt.Errorf("message length < 13")
	}
	splitstr:="\\"
	if f_windows==strings.ToLower(runtime.GOOS){

	}else {
		splitstr="/"
	}
	if len(msg.(string))==13{
		e.Fields.Put(f_msg,"")
	}else{
		msgstr:=msg.(string)[13:]
		e.Fields.Put(f_msg,msgstr)
	}

	//adf 前12个字符是时间
	timestr :=substring(msg.(string),0,12)

	hourstr:= substring(timestr,0,2)
	houri,err := strconv.Atoi(hourstr)

	minutestr:=substring(timestr,3,5)
	minutei,err := strconv.Atoi(minutestr)

	secondstr:=substring(timestr,6,8)
	secondi,err := strconv.Atoi(secondstr)

	millisecstr:=substring(timestr,9,12)
	milliseci,err := strconv.Atoi(millisecstr)

	source,err:=e.Fields.GetValue(f_source)
	if err!=nil{
		return fmt.Errorf("not existed field source")
	}

	sources:=strings.Split(source.(string),splitstr)

	//for _,s:=range sources{
	//	fmt.Println(s)
	//}
	lensources:=len(sources)
	if lensources<5{
		//解析时间字段
		return fmt.Errorf("length of sources(split by \\) < 5")
	}

	yearstr:=sources[lensources-2][0:4]
	yeari,err:=strconv.Atoi(yearstr)
	monthstr:=sources[lensources-2][4:6]
	monthi,err:=strconv.Atoi(monthstr)
	daystr:=strings.Split(sources[lensources-1],".")[0]
	dayi,err:=strconv.Atoi(daystr)

	//fmt.Println("$$$$sss")
	//fmt.Println(yeari)
	//fmt.Println(monthi)
	//fmt.Println(dayi)
	//fmt.Println("$$$$")

	t := time.Date(yeari, time.Month(monthi), dayi, houri, minutei, secondi, 0, time.Local)
	datetimeutc:=t.Unix()*1000+int64(milliseci)
	s:=t.Format("yyyy.MM.dd")
	fmt.Println(s)
	e.Fields.Put(f_logdate,s)
	e.Fields.Put(f_datetime,datetimeutc)
	return nil
}

func convertChar2Field(char byte) string{
	switch char{
	case 'c':return "CATEGORY"
	case 'C':return "CLASS"
	case 'd':return f_datetime//"DATE"
	case 'F':return "FILE"
	case 'l':return "LOCATION"
	case 'L':return  "LINE"
	case 'm':return  f_msg//"MESSAGE"
	case 'M':return "METHOD"
	case 'n':return  "LINE_SEPERATOR"
	case 'p':return f_loglevel//"PRIORITY"
	case 'r':return "ELAPSED"
	case 't':return "THREAD"
	case 'x':return "NDC"
	case 'X':return "MDC"
	default:
		return ""
	}
	return ""
}
func substring(source string, start int, end int) string {
	var r = []rune(source)
	length := len(r)

	if start < 0 || end > length || start > end {
		return ""
	}

	if start == 0 && end == length {
		return source
	}

	return string(r[start : end])
}


