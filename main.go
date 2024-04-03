package main

import (
	"fmt"
	"time"
)

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

type Task struct {
	id             int
	creationTime   string // время создания
	completionTime string // время выполнения
	result         []byte
}

func main() {
	taskCreator := func(taskChannel chan Task) {
		go func() {
			for {
				creationTime := time.Now().Format(time.RFC3339)
				if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
					creationTime = "Some error occurred"
				}
				taskChannel <- Task{id: int(time.Now().Unix()), creationTime: creationTime}
			}
		}()
	}

	taskWorker := func(task Task) Task {
		createdTime, _ := time.Parse(time.RFC3339, task.creationTime)
		if createdTime.After(time.Now().Add(-20 * time.Second)) {
			task.result = []byte("Task has been successfully completed")
		} else {
			task.result = []byte("Something went wrong")
		}
		task.completionTime = time.Now().Format(time.RFC3339Nano)

		time.Sleep(time.Millisecond * 150)

		return task
	}

	doneTasks := make(chan Task)
	undoneTasks := make(chan Task)

	taskSorter := func(task Task) {
		if string(task.result[14:]) == "successfully completed" {
			doneTasks <- task
		} else {
			undoneTasks <- task
		}
	}

	taskChannel := make(chan Task, 10)
	go taskCreator(taskChannel)

	go func() {
		for task := range taskChannel {
			task = taskWorker(task)
			go taskSorter(task)
		}
		close(taskChannel)
	}()

	results := make(map[int]Task)
	errors := []error{}

	go func() {
		for {
			select {
			case doneTask := <-doneTasks:
				results[doneTask.id] = doneTask
			case undoneTask := <-undoneTasks:
				errMsg := fmt.Errorf("Task id %d, creation time %s, error %s", undoneTask.id, undoneTask.creationTime, undoneTask.result)
				errors = append(errors, errMsg)
			}
		}
	}()

	time.Sleep(time.Second * 3)

	fmt.Println("Errors:")
	for _, err := range errors {
		fmt.Println(err)
	}

	fmt.Println("Done tasks:")
	for id := range results {
		fmt.Println(id)
	}
}
