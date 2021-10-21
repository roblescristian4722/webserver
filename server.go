package main

import (
    "errors"
    "fmt"
    "strconv"
    "net"
    "net/rpc"
    "net/http"
    "io/ioutil"
)

// RPC
type Server struct {
    Materias, Alumnos map[string] map[string] float64
}

type Args struct {
    Nombre, Materia string
    Cal float64
}

var serIns *Server

func printData(title string, m map[string]map[string]float64) {
    fmt.Println(title)
    for k, v := range m {
        fmt.Printf("    * %s:\n", k)
        for ki, vi := range v {
            fmt.Printf("        - %s: %f\n", ki, vi)
        }
    }
}

func (t *Server) add(student string, class string, grade float64) {
    if _, err := t.Alumnos[student]; err {
        t.Alumnos[student][class] = grade
    } else {
        fmt.Printf("[Nuevo alumno añadido: %s]\n", student)
        m := make(map[string] float64)
        m[class] = grade
        t.Alumnos[student] = m
    }
    if _, err := t.Materias[class]; err {
        t.Materias[class][student] = grade
    } else {
        fmt.Printf("[Nueva materia añadida: %s]\n", class)
        n := make( map[string] float64 )
        n[student] = grade
        t.Materias[class] = n
    }
}

func (t *Server) AddGrade(args Args, reply *int) error {
    fmt.Println()
    t.add(args.Nombre, args.Materia, args.Cal)
    printData("Alumnos: ", t.Alumnos)
    printData("Materias: ", t.Materias)
    fmt.Println("-----------------------------------------")
    return nil
}

func (t *Server) studentMean(name string) float64 {
    var res float64
    var n float64
    for _, v := range t.Alumnos[name] {
        res += v
        n++
    }
    res /= n
    return res
}

func (t *Server) generalMean() float64 {
    var res float64
    var n float64
    for k, _ := range t.Alumnos {
        res += t.studentMean(k)
        n++
    }
    res /= n
    return res
}

func (t *Server) StudentMean(args Args, reply *float64) error {
    if _, err := t.Alumnos[args.Nombre]; !err {
        return errors.New("El usuario " + args.Nombre + " no fue registrado con anterioridad")
    }
    (*reply) = t.studentMean(args.Nombre)
    return nil
}

func (t *Server) GeneralMean(args Args, reply *float64) error {
    if len(t.Alumnos) == 0 {
        return errors.New("No hay alumnos registrados")
    }
    (*reply) = t.generalMean()
    return nil
}

func (t *Server) ClassMean(args Args, reply *float64) error  {
    if _, err := t.Materias[args.Materia]; !err {
        return errors.New("La materia " + args.Materia + " no fue registrada con anterioridad")
    }
    var res float64
    var n float64
    for _, v := range t.Materias[args.Materia] {
        res += v
        n++
    }
    res /= n
    (*reply) = res
    return nil
}

func handleRpc(s *Server) {
    rpc.Register(s)
    rpc.HandleHTTP()
    ln, err := net.Listen("tcp", ":9999")
    if err != nil {
        fmt.Println(err)
        return
    }
    for {
        c, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        go rpc.ServeConn(c)
    }
}

// HTTP
func readHTML(fileName string) string {
    html, _ := ioutil.ReadFile(fileName)
    return string(html)
}

func generalMean(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type", "text/html")
    if len((*serIns).Alumnos) == 0 {
        fmt.Fprintln(res, readHTML("./generalMeanError.html"))
    } else {
        fmt.Fprintf(res, readHTML("./generalMeanError.html"), (*serIns).generalMean())
    }
}

func add(res http.ResponseWriter, req *http.Request) {
    if err := req.ParseForm(); err != nil {
        fmt.Fprintf(res, "ParseForm() error %v", err)
        return
    }
    res.Header().Set("Content-Type", "text/html")
    fmt.Fprintf(res, readHTML("./addForm.html"))
}

func registry(res http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case "POST":
        cal, _ := strconv.ParseFloat(req.FormValue("cal"), 64)
        (*serIns).add(req.FormValue("alu"), req.FormValue("mat"), cal)
        res.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(res,
                    readHTML("./addRes.html"),
                    req.FormValue("alu"),
                    req.FormValue("mat"),
                    cal)
    case "GET":

    }
}

func studentMean(res http.ResponseWriter, req *http.Request) {

}

func classMean(res http.ResponseWriter, req *http.Request) {
}

func main() {
    arith := new(Server)
    arith.Alumnos = make(map[string]map[string]float64)
    arith.Materias = make(map[string]map[string]float64)
    serIns = arith
    // Se gestionan las peticiones rpc en una goroutine
    fmt.Println("Corriendo servidor rpc...")
    go handleRpc(arith)
    fmt.Println("Corriendo servidor http...")
    // Se gestionan las peticiones http en el flujo principal
    http.HandleFunc("/agregar", add)
    http.HandleFunc("/registros", registry)
    http.HandleFunc("/promedio_alumno", studentMean)
    http.HandleFunc("/promedio_general", generalMean)
    http.HandleFunc("/promedio_materia", classMean)
    http.ListenAndServe(":9000", nil)
}
