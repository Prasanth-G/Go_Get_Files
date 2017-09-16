package main
import(
	"fmt"
	"log"
	"net/http"
	"os"
	"io"
	"net"
	"path"
	"flag"
)

var progAddr string
var saveTo = "."

func searchForConn() map[string]string{

	ifaces, _ := net.Interfaces()
	l := make(map[string]string)
	for _, i := range ifaces {
		if i.Flags.String() == net.FlagUp.String() + "|" + net.FlagBroadcast.String() + "|" + net.FlagMulticast.String(){
			addrs, _ := i.Addrs()
			addr := addrs[len(addrs) - 1].String()
			l[i.Name] = addr[0:len(addr) - 3]
		}
	}
	return l
}

func receiveFile(writer http.ResponseWriter, request *http.Request){

	switch request.Method{
		case "POST":
			request.ParseMultipartForm(100000)
			mpf := request.MultipartForm

			files := mpf.File["files"]
			for index, _ := range(files){

				log.Println("receiving", files[index].Filename)
				file, err := files[index].Open()
				if (err != nil){
					log.Println("Error Opening file", err)
				}
				defer file.Close()
				dest, err := os.Create(path.Join(saveTo, files[index].Filename))
				if (err != nil){
					log.Println("Error Opening file", err)
				}
				defer dest.Close()
				writer.WriteHeader(200)
				writer.Write([]byte("<h1>Thank you</h1><h5>I received your file[s]</h5>"))
				if _, err := io.Copy(dest, file); err != nil {
					log.Println("Error copying", err)
				}
				log.Println(files[index].Filename, "received")
			}

		default:
			http.Redirect(writer, request, "/", 301)
			return
	}
}

func createStaticDir(){
	var content = `<!DOCTYPE html>
	<html lang="en">
	  <head>
		<title>File Upload</title>
	  </head>
	  <body>
		  <h1>File Upload</h1>
		  <form method="post" action="/upload" enctype="multipart/form-data">
				<input type="file" name="files" multiple="multiple">
				<input type="submit" value="Upload">
		  </form>
	  </body>
	</html>`
	err := os.Mkdir("static", os.ModePerm)
	if err != nil{
		log.Println("Error MKDIR", err)
	}
	file, err := os.Create(path.Join("static", "index.html"))
	if err != nil{
		log.Println("Error Creating index.html", err)
	}
	file.Write([]byte(content))
	file.Close()
}

func main(){
	
	var port string
	flag.StringVar(&port, "port", ":8000", "Port Number [range : 0 - 65536]")
	flag.StringVar(&saveTo, "saveto", ".", "Relative or Absolute Path to store the Received Files")
	flag.Parse()

	if _, err := os.Stat("./static"); os.IsNotExist(err){
		createStaticDir()
	}

	l := make(map[string]string)
	for l = searchForConn(); len(l) == 0;{
		fmt.Println("No Connections available... Refresh ? (yes/ no)")
		var s string
		fmt.Scan(&s)
		if s == "yes" || s == "y"{
			l = searchForConn()
		}else{
			os.Exit(0);
		}
	}
	var i = 1
	options := make([]string, 5)
	if len(l) > 1{
		fmt.Println("More than one Connections are available..\nChoose any one ")
		count := 0
		for key, value := range(l){
			fmt.Println(count + 1, ". ", key, " - ", value)
			options[count] = key
			count = count + 1
		}
		fmt.Scan(&i)

	}else{
		for key := range(l){
			options[0] = key
		}
	}
	progAddr = l[options[i - 1]] + port
	fmt.Println("Get files from your Friends by Visiting the following IP address from your Friend's browser")
	fmt.Println(progAddr)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static")))
	defer http.ListenAndServe(port, mux)
	mux.HandleFunc("/upload", receiveFile)

}
