package main

import (
	"fmt"
	"sort"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	var in chan interface{}
	wg := &sync.WaitGroup{}

	for _, cmdInstance := range cmds {
		out := make(chan interface{})
		wg.Add(1)
		go func(in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			cmdInstance(in, out)
		}(in, out)
		in = out
	}

	wg.Wait()

}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	users := map[string]struct{}{}
	asyncGetUser := func(email string) {
		defer wg.Done()
		user := GetUser(email)
		mu.Lock()
		if _, exists := users[user.Email]; exists {
			mu.Unlock()
			return
		}
		users[user.Email] = struct{}{}
		mu.Unlock()
		out <- user
	}
	for fromChannel := range in {
		email := fromChannel.(string)
		wg.Add(1)
		go asyncGetUser(email)
	}
	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	wg := &sync.WaitGroup{}
	processUserMessages := func(batch []User) {
		defer wg.Done()
		messages, _ := GetMessages(batch...)
		for _, message := range messages {
			out <- message
		}
	}
	batch := make([]User, 0, GetMessagesMaxUsersBatch)
	for userFromChannel := range in {
		user := userFromChannel.(User)
		batch = append(batch, user)
		if len(batch) == GetMessagesMaxUsersBatch {
			wg.Add(1)
			currBatch := append([]User(nil), batch...)
			go processUserMessages(currBatch)
			batch = []User{}
		}
	}
	if len(batch) != 0 {
		wg.Add(1)
		go processUserMessages(batch)
	}

	wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
	queue := make(chan MsgID)
	wg := &sync.WaitGroup{}

	startWorker := func() {
		defer wg.Done()
		for msgID := range queue {
			res, _ := HasSpam(msgID)
			out <- MsgData{ID: msgID, HasSpam: res}
		}
	}

	for range HasSpamMaxAsyncRequests {
		wg.Add(1)
		go startWorker()
	}

	for msgID := range in {
		coercedID := msgID.(MsgID)
		queue <- coercedID
	}

	close(queue)
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	var messages []MsgData
	for msg := range in {
		msgData := msg.(MsgData)
		messages = append(messages, msgData)
	}

	sort.Slice(messages, func(i, j int) bool {
		if messages[i].HasSpam && messages[j].HasSpam || !messages[i].HasSpam && !messages[j].HasSpam {
			return messages[i].ID < messages[j].ID
		}

		return messages[i].HasSpam
	})

	for _, message := range messages {
		out <- fmt.Sprintf("%v %v", message.HasSpam, message.ID)
	}
}
