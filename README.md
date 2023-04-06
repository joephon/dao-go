# My toy language

### build
```sh
go build main.go
```

### run repl
```sh
./main
```

### run source file
```sh
./main <your/source/file>
./main sample/closure.go
```


### commands
```sh
./main -h
./main -v
```

### string
```go
var str string = "Hello"
var wd string = "World!"
var hi = str + " " + wd + " "
// => "Hello World! "

hi * 2
echo(hi)
// => "Hello World! Hello World! "
len(hi)
// => 26
```

### fibonacc
```go
func fibonacc(x int) {
    if x == 0 {
        0
    } else if x == 1 {
        1
    } else {
        fibonacc(x - 1) + fibonacc(x - 2)
    }
}

fibonacc(35)
```

### closure
```go
func closure() {
    var a = 0
    func(x int) {
        a = a + x
        a
    }
}

var c = closure()
c(1)
c(1)
c(1)
c(1)
c(1)
```

### loop

```go
func a() {
    for var a = 0; a = a + 1; a < 6 {
        puts("a is: ", a)
    }
}

func b() {
    var b = 0
    for {
        puts("b is: ", b)
        b = b + 1
        
        if b > 5 {
            return nil
        }
        
    }
}

func c() {
    var c = 0
    for c = c + 1;  c < 6 {
        puts("c is: ", c)
        
    }
}

a(); echo("-------")
b(); echo("-------")
c()
```