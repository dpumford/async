## Async

This package simplifies the process of running code concurrently. It codifies a common `WaitGroup` + channels pattern to reduce the amount of code required to start an arbitrary  number of concurrent functions and wait for their results. The functions use generics to keep our code honest at compile time.

```golang
func main() {
    multiplierQueue := function_queue.NewFunctionQueue[int, int](3)

    for i := 0; i < 10; i++ {
        multiplierQueue.Run(func(num int) (int, error) {
            return num * num, nil
        }, i)
    }

    results := multiplierQueue.Wait()

    for _, v := range results {
        if v.Err != nil {
            panic(v.Err)
        }

        fmt.Println(v.R)
    }
}
```

A larger example service can be found at [`./examples/cat_facts`](./examples/cat_facts/main.go).