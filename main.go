package main

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

func main() {
	args := os.Args

	arg_count := len(args)

	if 2 > arg_count || arg_count > 3 {
		fmt.Printf("Usage: %s <FLAG> <FILE NAME>\n", args[0])
		os.Exit(1)
	}

	has_flag := arg_count == 3
	if has_flag && args[1] != "-a" {
		fmt.Printf("%s is not a valid flag\n", args[1])
		os.Exit(2)
	}
	file_to_find := args[1+(arg_count-2)]

	file_name, extension := split_file_name(file_to_find)

	ch := make(chan FileMatch, 100)
	var wg sync.WaitGroup

	start_time := time.Now()
	if has_flag {
		fmt.Printf("Starting search for %s.%s in all files\n", file_name, extension)
		search_all_drives(file_name, extension, ch, &wg)
	} else {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory:", err)
			os.Exit(3)
		}
		fmt.Printf("Starting search for %s.%s in %s\n", file_name, extension, dir)
		wg.Add(1)
		go search_folder(dir, file_name, extension, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for result := range ch {
		percent := int(math.Round(float64(result.match * 100)))
		fmt.Printf("%s.%s is a %d%% match\n", result.name, result.extension, percent)
	}
	fmt.Println("Done in: ", time.Since(start_time))

	os.Exit(0)
}
