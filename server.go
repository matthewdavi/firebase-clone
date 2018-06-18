package main

import (
    "net/http"
    "log"
    "os"
    "strings"
    "io/ioutil"
    "time"
    "html/template"
    "github.com/gorilla/websocket"
)

var (
    upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    }
)

const baseHtml = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>{{.Title}}</title>
    </head>
    <body>
        Hello {{.Name}}!!!
        Input:
        <textarea id="content"></textarea>
        Output:
        <p id="server-output"></p>
    </body>
    <script>
        // FIXME: Should wait for onload and onmessage in parallel
        window.onload = function() {
            let wsConnection = new WebSocket("ws://localhost{{.Port}}/ws?file=f1");
            wsConnection.onmessage = function(evt) {
                console.log(evt)
            }

            let textArea = document.getElementById("content")
            textArea.onkeyup = function() {
                wsConnection.send(textArea.value)
            }

            wsConnection.onmessage = function(evt) {
                document.getElementById("server-output").innerHTML = evt.data
            }
        }
    </script>
</html>
`

func reader(conn *websocket.Conn, slug []string) {
    for {
        _, p, err := conn.ReadMessage()
        if err != nil {
            log.Println(err)
            return
        }

        err = ioutil.WriteFile("./userfiles/" + strings.Join(slug, ""), p, 0660)
        if err != nil {
            log.Println(err)
            return
        }
    }
}

func writer(conn *websocket.Conn, slug []string) {
    // HACK: This is just for testing. Obviously should have some kind of subscription.
    // But we are writing to the file system as a temporary hack anyway, so not worth doing this correctly.
    for range time.Tick(time.Millisecond * 500) {
        p, err := ioutil.ReadFile("./userfiles/" + strings.Join(slug, ""))
        if err != nil {
            log.Println(err)
            return
        }
        conn.WriteMessage(websocket.TextMessage, p)
    }
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.New("").Parse(baseHtml)
    if err != nil {
        log.Println(err)
        return
    }

    var data = struct{
        Name string
        Title string
        Port string
    }{"Michael", "Fake Firebase", os.Getenv("PORT")}

    tmpl.Execute(w, data)
}

func serveWs(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    defer conn.Close()
    if err != nil {
        log.Println(err)
        return
    }

    file := r.URL.Query()["file"]
    if len(file) == 0 {
        log.Println("No file specified")
        return
    }

    go writer(conn, file)
    reader(conn, file)
}

func main() {
    http.HandleFunc("/", serveStatic)
    http.HandleFunc("/ws", serveWs)

    log.Fatal(http.ListenAndServe(os.Getenv("PORT"), nil))
}
