package main

import (
	"fmt"
	"naloga2/socialNetwork"
	"strings"
	"sync"
	"time"
	"unicode"
)

var rw_lock sync.RWMutex

func cleanUp(text string) string {
	var builder strings.Builder
	for _, c := range text {
		if unicode.IsLetter(c) || unicode.IsSpace(c) {
			builder.WriteRune(unicode.ToLower(c))
		} else {
			builder.WriteRune(' ')
		}
	}
	return builder.String()
}

func processTask(task socialNetwork.Task, inverted_index map[string][]uint64) {

	all_words := make(map[string]int)
	clean_words := strings.Fields(cleanUp(task.Data))

	for _, word := range clean_words {
		if len(word) >= 4 {
			all_words[word]++
		}
	}

	rw_lock.Lock()
	for word := range all_words {
		inverted_index[word] = append(inverted_index[word], task.Id)
	}
	rw_lock.Unlock()
}

func worker(task_queue <-chan socialNetwork.Task, inverted_index map[string][]uint64, quit <-chan bool, wg_workers *sync.WaitGroup) {
	defer wg_workers.Done()

	for {
		select {
		case task := <-task_queue:
			processTask(task, inverted_index)
		case <-quit:
			return
		}
	}
}

// za samo en test z enim delayom

// func supervisor(producer *socialNetwork.Q, inverted_index map[string][]uint64, stop_supervisor chan bool) {
// 	min_workers := 1
// 	maxWorkers := 100
// 	//check_interval := time.Second / 100
// 	var active_workers int = (min_workers + maxWorkers) / 2

// 	task_queue := producer.TaskChan
// 	quit := make(chan bool)

// 	var wg_workers sync.WaitGroup
// 	for i := 0; i < active_workers; i++ {
// 		wg_workers.Add(1)
// 		go worker(task_queue, inverted_index, quit, &wg_workers)
// 	}

// 	for {
// 		//time.Sleep(check_interval)
// 		select {
// 		case <-stop_supervisor:

// 			for len(task_queue) > 0 {
// 			}

// 			for i := 0; i < active_workers; i++ {
// 				quit <- true
// 			}

// 			wg_workers.Wait()
// 			close(quit)
// 			return

// 		default:
// 			queue_len := len(task_queue)

// 			if queue_len > 8000 && active_workers < maxWorkers {
// 				active_workers++
// 				wg_workers.Add(1)
// 				go worker(task_queue, inverted_index, quit, &wg_workers)
// 				fmt.Println("+, total:", active_workers)
// 			} else if queue_len < 2000 && active_workers > min_workers {
// 				quit <- true
// 				active_workers--
// 				fmt.Println("-, total:", active_workers)
// 			}
// 		}
// 	}
// }

// func main() {
// 	var inverted_index = make(map[string][]uint64)
// 	var producer socialNetwork.Q
// 	delay := 10000
// 	stop_supervisor := make(chan bool)

// 	producer.New(delay)
// 	go producer.Run()

// 	start := time.Now()
// 	go supervisor(&producer, inverted_index, stop_supervisor)
// 	time.Sleep(time.Second * 5)

// 	producer.Stop()
// 	stop_supervisor <- true
// 	close(stop_supervisor)

// 	for !producer.QueueEmpty() {
// 	}

// 	elapsed := time.Since(start)
// 	// Izpišemo število generiranih zahtevkov na sekundo
// 	fmt.Printf("Processing rate: %f MReqs/s\n", float64(producer.N)/float64(elapsed.Seconds())/1000000.0)
// 	// Izpišemo povprečno dolžino vrste v čakalnici
// 	fmt.Printf("Average queue length: %.2f %%\n", producer.GetAverageQueueLength())
// 	// Izpišemo največjo dolžino vrste v čakalnici
// 	fmt.Printf("Max queue length %.2f %%\n", producer.GetMaxQueueLength())
// }



//testiranje za razlicne delaye

func supervisor(producer *socialNetwork.Q, inverted_index map[string][]uint64, stop_supervisor chan bool) {
	min_workers := 1
	maxWorkers := 100
	//check_interval := time.Second / 100
	var active_workers int = (min_workers + maxWorkers) / 2

	task_queue := producer.TaskChan
	quit := make(chan bool)

	var wg_workers sync.WaitGroup
	for i := 0; i < active_workers; i++ {
		wg_workers.Add(1)
		go worker(task_queue, inverted_index, quit, &wg_workers)
	}

	for {
		//time.Sleep(check_interval)
		select {
		case <-stop_supervisor:

			for len(task_queue) > 0 {
			}

			for i := 0; i < active_workers; i++ {
				quit <- true
			}

			wg_workers.Wait()
			close(quit)
			return

		default:
			queue_len := len(task_queue)

			if queue_len > 8000 && active_workers < maxWorkers {
				active_workers++
				wg_workers.Add(1)
				go worker(task_queue, inverted_index, quit, &wg_workers)
			} else if queue_len < 2000 && active_workers > min_workers {
				quit <- true
				active_workers--
			}
		}
	}
}

func main() {

	for delay:=1000; delay<10001; delay+=1000{

		var inverted_index = make(map[string][]uint64)
		var producer socialNetwork.Q
		stop_supervisor := make(chan bool)
	
		producer.New(delay)
		go producer.Run()
	
		start := time.Now()
		go supervisor(&producer, inverted_index, stop_supervisor)
		time.Sleep(time.Second * 5)
	
		producer.Stop()
		stop_supervisor <- true
		close(stop_supervisor)
	
		for !producer.QueueEmpty() {
		}
	
		elapsed := time.Since(start)

		fmt.Printf("Delay: %d\n", delay)
		// Izpišemo število generiranih zahtevkov na sekundo
		fmt.Printf("Processing rate: %f MReqs/s\n", float64(producer.N)/float64(elapsed.Seconds())/1000000.0)
		// Izpišemo povprečno dolžino vrste v čakalnici
		fmt.Printf("Average queue length: %.2f %%\n", producer.GetAverageQueueLength())
		// Izpišemo največjo dolžino vrste v čakalnici
		fmt.Printf("Max queue length %.2f %%\n", producer.GetMaxQueueLength())
		fmt.Println();
	}

}


//stara koda brez timeouta:

// func supervisor(producer *socialNetwork.Q, inverted_index map[string][]uint64, wg *sync.WaitGroup) {
// 	defer wg.Done()

// 	min_workers := 1
// 	maxWorkers := 100
// 	check_interval := time.Second / 100
// 	var active_workers int = min_workers

// 	task_queue := producer.TaskChan
// 	quit := make(chan bool)

// 	var wg_workers sync.WaitGroup
// 	for i := 0; i < active_workers; i++ {
// 		wg_workers.Add(1)
// 		go worker(task_queue, inverted_index, quit, &wg_workers)
// 	}

// 	for {
// 		time.Sleep(check_interval)
// 		queue_len := len(task_queue)

// 		if queue_len > 8000 && active_workers < maxWorkers {
// 			active_workers++
// 			wg_workers.Add(1)
// 			go worker(task_queue, inverted_index, quit, &wg_workers)
// 			fmt.Println("Added worker, total:", active_workers)
// 		} else if queue_len < 2000 && active_workers > min_workers {
// 			quit <- true
// 			active_workers--
// 			fmt.Println("Removed worker, total:", active_workers)
// 		}

// 		if producer.QueueEmpty() && active_workers == min_workers {
// 			for i := 0; i < min_workers; i++ {
// 				quit <- true
// 			}
// 			wg_workers.Wait()
// 			close(quit)
// 			return
// 		}
// 	}
// }

// func main() {
// 	var inverted_index = make(map[string][]uint64)
// 	var producer socialNetwork.Q
// 	delay := 7000

// 	producer.New(delay)
// 	go producer.Run()

// 	var wg sync.WaitGroup
// 	wg.Add(1)

// 	start := time.Now()
// 	go supervisor(&producer, inverted_index, &wg)

// 	wg.Wait()
// 	for !producer.QueueEmpty() {}
// 	producer.Stop()

// 	elapsed := time.Since(start)
// 	// Izpišemo število generiranih zahtevkov na sekundo
// 	fmt.Printf("Processing rate: %f MReqs/s\n", float64(producer.N)/float64(elapsed.Seconds())/1000000.0)
// 	// Izpišemo povprečno dolžino vrste v čakalnici
// 	fmt.Printf("Average queue length: %.2f %%\n", producer.GetAverageQueueLength())
// 	// Izpišemo največjo dolžino vrste v čakalnici
// 	fmt.Printf("Max queue length %.2f %%\n", producer.GetMaxQueueLength())
// }
