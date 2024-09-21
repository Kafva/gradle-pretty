package main

import (
    "bufio"
    "flag"
    "fmt"
    "os"
    "strings"
    "time"
)

type GradleTask struct {
    Name string
    Failed bool
}

type GradleError struct {
    Location string
    Desc string
}

func die(fmtStr string, args... any) {
    fmt.Printf(fmtStr, args ...)
    os.Exit(1)
}

func taskLog(task GradleTask) {
    var result string
    if task.Failed {
        result = " \033[91m\033[0m"
    } else {
        result = ""
    }
    msg := fmt.Sprintf("\033[93m▸\033[0m %s%s", task.Name, result)
    print("\r\033[2K")  // Clear line
    println(msg)        // Print with newline
}

func buildOk(tasks []GradleTask, errors []GradleError) bool {
    if len(tasks) == 0 || len(errors) > 0 {
        return false
    }

    for _,task := range tasks {
        if task.Failed {
            return false
        }
    }

    return true
}

func parseBuildLog(noLogfile bool, logfile string) (
    tasks []GradleTask,
    errors []GradleError,
    timeTaken int64) {
    var f *os.File
    var err error
    if !noLogfile {
        f, err = os.OpenFile(logfile,
                             os.O_TRUNC | os.O_CREATE | os.O_WRONLY, 0644)
        if err != nil {
            die("Error opening %s: %s\n", logfile, err)
        }
        defer f.Close()
    }
    cwd, err := os.Getwd()
    if err != nil {
        die("Error reading current working directory: %s\n", err)
    }
    cwd = "file://" + cwd + "/"

    scanner := bufio.NewScanner(os.Stdin)

    startTime := time.Now().Unix()
    for scanner.Scan() {
        rawline := scanner.Text()
        line := strings.TrimSpace(rawline)

        if strings.HasPrefix(line, "> Task") {
            spl := strings.Split(line, " ")
            if len(spl) < 4 {
                continue
            }
            task := GradleTask { spl[2], spl[3] == "FAILED" }
            tasks = append(tasks, task)
            if len(tasks) > 1 {
                // After the first log line, always move back up one line
                // before printing anew
                print("\033[A")
            }
            taskLog(task)

        } else if strings.HasPrefix(line, "e:") {
            // Source code errors have an 'e:' prefix
            spl := strings.Split(line, " ")
            if len(spl) < 2 {
                continue
            }
            location, _ := strings.CutPrefix(spl[1], cwd)
            desc := strings.Join(spl[2:], " ")
            err := GradleError { location, desc }
            errors = append(errors, err)

        }

        if !noLogfile {
            f.WriteString(rawline + "\n")
        }
    }
    endTime := time.Now().Unix()
    timeTaken = endTime - startTime

    return tasks, errors, timeTaken
}

func main() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "Prettify the gradle build log:\n")
        fmt.Fprintf(os.Stderr, "  ./gradlew build 2>&1 | %s\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "\nOPTIONS:\n")
        flag.PrintDefaults()
    }
    logfile := flag.String("l", "build.log", "Path to save complete build log in")
    noLogfile := flag.Bool("N", false, "Do not save a copy of the complete build log")
    flag.Parse()

    tasks, errors, timeTaken := parseBuildLog(*noLogfile, *logfile)

    if buildOk(tasks, errors) {
        fmt.Printf("\033[92mBUILD SUCCESSFUL\033[0m in %ds\n", timeTaken)
        os.Exit(0)

    } else {
        for _,err := range errors {
            fmt.Printf("%s: %s\n", err.Location, err.Desc)
        }

        fmt.Printf("\033[91mBUILD FAILED\033[0m in %ds\n", timeTaken)
        if len(tasks) == 0 {
            println("No tasks completed")
        }
        if ! *noLogfile {
            fmt.Printf("See %s for more information\n", *logfile)
        }
        os.Exit(1)
    }
}
