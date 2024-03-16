package main

import (
    "net/http"
    "log"
    "fmt"
	"io"
	"os"
	"time"
	"crypto/rand"
	rndm "math/rand"
	"crypto/md5"
	"path/filepath"

    "github.com/julienschmidt/httprouter"
)


const MAX_UPLOAD_SIZE = 10 * 1024 * 1024 // 10 mb

//const posterPath = "./posters"
const videoPath = "./videos"
const imagePath = "./images"


func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/", Index)
	
	static := httprouter.New()
	static.ServeFiles("/image/*filepath", http.Dir(imagePath))
	static.ServeFiles("/video/*filepath", http.Dir(videoPath))
//	static.ServeFiles("/poster/*filepath", http.Dir(posterPath))
	router.NotFound = static


	log.Println("Starting Server")
    log.Fatal(http.ListenAndServe(":4173", router))
}


func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	enableCors(&w)
	switch r.Method {
		case "OPTIONS" : {
			fmt.Fprintf(w,"200")
			return
		}
		case "POST" : {
			r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
			if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
				http.Error(w, "The uploaded file is too big. Please choose an file that's less than 10MB in size", http.StatusBadRequest)
			}
			fmt.Println("Key :  ", r.FormValue("key"))
			//~ fmt.Println(r.Body)
			ffile, fileHeader, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			defer ffile.Close()
			var fileEndings string
			var folderpath string
			var fileName string
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer file.Close()
			// Get and print outfile size
			fileSize := fileHeader.Size
		//	FileTitle := strings.Split(fileHeader.Filename, ".")[0]
			fmt.Printf("File size (bytes): %v\n", fileSize)
			// validate file size
			if fileSize > MAX_UPLOAD_SIZE {
				renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			}
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				renderError(w, "INVALID_FILE"+http.DetectContentType(fileBytes), http.StatusBadRequest)
			}

			//~ // check file type, detectcontenttype only needs the first 512 bytes
			detectedFileType := http.DetectContentType(fileBytes)
			switch detectedFileType {
			case "video/mp4":
				fileEndings = ".mp4"
				folderpath = "./videos"
				break
			case "video/webm":
				fileEndings = ".webm"
				folderpath = "./videos"
				break
			case "image/gif":
				fileEndings = ".gif"
				folderpath = "./images"
				break
			case "image/png":
				fileEndings = ".png"
				folderpath = "./images"
				break
			case "image/webp":
				fileEndings = ".webp"
				folderpath = "./images"
				break
			case "image/jpg":
				fileEndings = ".jpg"
				folderpath = "./images"
				break
			case "image/jpeg":
				fileEndings = ".jpeg"
				folderpath = "./images"
				break
			default:
				renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			}
			fileName = r.FormValue("key")
			fmt.Println("fileName : ",fileName)
			if err != nil {
				renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			}
			//~ newFileName := fileName + fileEndings
			//~ newPath := filepath.Join(uploadPath, newFileName)
			//~ newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), fileEndings)
			newFileName := fmt.Sprintf("%d%s", GenerateFileName(16), fileEndings)
			fmt.Println(newFileName)
			newPath := filepath.Join(folderpath, newFileName)
			//~ newPath := fmt.Sprintf("./images/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
			//~ fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

			fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
			fmt.Printf("File Size: %+v\n", fileHeader.Size)
			fmt.Printf("MIME Header: %+v\n", fileHeader.Header)
			// write file
			newFile, err := os.Create(newPath)
			if err != nil {
				renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			}
			defer newFile.Close() // idempotent, okay to call twice
			if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
				renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			}
			fmt.Println(r.Body)
			fmt.Fprintf(w, newFileName)
		}
		default: {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}


func GenerateFileName(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ")

    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rndm.Intn(len(letters))]
    }
    return string(b)
}

func init() {
    rndm.Seed(time.Now().UnixNano())
}


//~ func sendImageAsHTML(w http.ResponseWriter, r *http.Request, a string) {
	//~ fmt.Fprintf(w,a)
//~ }

//~ func sendImageAsAttachment(w http.ResponseWriter, r *http.Request) {
    //~ buf, err := os.ReadFile("F46ZPQ0bQAACFYs.jpg")
    //~ if err != nil {
        //~ log.Fatal(err)
    //~ }
    //~ w.Header().Set("Content-Type", "image/jpeg")
    //~ w.Header().Set("Content-Disposition", `attachment;filename="F46ZPQ0bQAACFYs.jpg"`)
    //~ w.Write(buf)
//~ }

//~ func sendImageAsBytes(w http.ResponseWriter, r *http.Request, a string) {
    //~ buf, err := os.ReadFile("./uploads/"+a)
    //~ if err != nil {
        //~ log.Fatal(err)
    //~ }
    //~ w.Header().Set("Content-Type", "image/jpeg")
    //~ w.Write(buf)
//~ }

//~ func DeleteEdit(filename string) {
	//~ time.Sleep(10 * time.Second)  
	//~ os.Remove(filename)
	//~ fmt.Println(filename,"Deleted")
//~ }


func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

//curl -F "file=@D:\Garbage\Blackpink\JisooCheeks[youtube@ouSU8JvC4vg]-3.webm" localhost:4000/upload/

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func rndmToken(len int) int64 {
	b := make([]byte, len)
	n, _ := rand.Read(b)
	return int64(n)
}

//~ func XHRrespond(w http.ResponseWriter, message string) {
	//~ fmt.Fprintf(w, message)
//~ }

//~ func EncrypIt(strToHash string) string {
	//~ data := []byte(strToHash)
	//~ return fmt.Sprintf("%x", md5.Sum(data))
//~ }

//~ func SessionVerify(sessionKey string) string {
	//~ return fmt.Sprintf(sessionKey)
//~ }

func Ignore(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "favicon.png")
}

