package main

// сюда писать код

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"github.com/uber/jaeger-client-go/crossdock/log"
)

var (
	// @BotFather в телеграме даст вам это
	BotToken = "7144449575:AAFdzhd4el8x5PoQH5H7xYnybjnz4xfDdoM"

	// урл выдаст вам игрок или хероку
	WebhookURL = "https://a415522c8bde5c4d-45-90-58-29.serveousercontent.com"
)

type User struct {
	ID       int
	Username string
}

type Task struct {
	ID       int64
	Name     string
	Assigner User
	Assignee User
	Done     bool
}

type TaskStore struct {
	mu     *sync.RWMutex
	tasks  []Task
	nextID int64
}

var taskStore = TaskStore{
	mu:     &sync.RWMutex{},
	nextID: 1,
}

func assign(ctx context.Context, taskID int) error {
	var taskToAssign *Task
	taskStore.mu.Lock()
	for idx, task := range taskStore.tasks {
		if task.ID == int64(taskID) {
			taskToAssign = &taskStore.tasks[idx]
		}
	}

	if taskToAssign == nil {
		taskStore.mu.Unlock()
		return errors.New("no such task")
	}

	bot := ctx.Value("bot").(*tgbotapi.BotAPI)
	chat := ctx.Value("chat").(*tgbotapi.Chat)

	var notifyID int64
	if taskToAssign.Assignee.ID != 0 {
		notifyID = int64(taskToAssign.Assignee.ID)
	} else {
		notifyID = int64(taskToAssign.Assigner.ID)
	}

	taskToAssign.Assignee = User{
		ID:       int(chat.ID),
		Username: chat.UserName,
	}
	taskStore.mu.Unlock()
	var err error
	if chat.ID != int64(taskToAssign.Assigner.ID) {
		message := tgbotapi.NewMessage(
			notifyID,
			fmt.Sprintf(`Задача "%s" назначена на @%s`, taskToAssign.Name, chat.UserName),
		)
		_, err = bot.Send(message)
		if err != nil {
			return err
		}
	}

	assigneeMessage := tgbotapi.NewMessage(
		chat.ID,
		fmt.Sprintf(`Задача "%s" назначена на вас`, taskToAssign.Name),
	)
	_, err = bot.Send(assigneeMessage)
	return err
}

func unassign(ctx context.Context, taskID int) error {
	var taskToAssign *Task
	taskStore.mu.Lock()
	for idx, task := range taskStore.tasks {
		if task.ID == int64(taskID) {
			taskToAssign = &taskStore.tasks[idx]
		}
	}

	if taskToAssign == nil {
		taskStore.mu.Unlock()
		return errors.New("no such task")
	}

	bot := ctx.Value("bot").(*tgbotapi.BotAPI)
	chat := ctx.Value("chat").(*tgbotapi.Chat)
	var err error
	if taskToAssign.Assignee.ID != int(chat.ID) {
		taskStore.mu.Unlock()
		message := tgbotapi.NewMessage(chat.ID, "Задача не на вас")
		_, err = bot.Send(message)
		if err != nil {
			return err
		}
	} else {
		taskToAssign.Assignee = User{}
		taskStore.mu.Unlock()

		_, err := bot.Send(tgbotapi.NewMessage(chat.ID, "Принято"))
		if err != nil {
			return err
		}
		_, err = bot.Send(
			tgbotapi.NewMessage(int64(taskToAssign.Assigner.ID),
				fmt.Sprintf(`Задача "%s" осталась без исполнителя`, taskToAssign.Name),
			))
	}

	return err
}

func resolve(ctx context.Context, taskID int) error {
	var taskToResolve Task
	taskStore.mu.Lock()
	for idx, task := range taskStore.tasks {
		if task.ID == int64(taskID) {
			taskToResolve = taskStore.tasks[idx]
			taskStore.tasks = append(taskStore.tasks[:idx], taskStore.tasks[idx+1:]...)
		}
	}
	taskStore.mu.Unlock()

	if taskToResolve.ID == 0 {
		return errors.New("task doesn't exist")
	}

	bot := ctx.Value("bot").(*tgbotapi.BotAPI)
	chat := ctx.Value("chat").(*tgbotapi.Chat)
	_, err := bot.Send(tgbotapi.NewMessage(chat.ID, fmt.Sprintf(`Задача "%s" выполнена`, taskToResolve.Name)))
	if err != nil {
		return err
	}
	if chat.ID != int64(taskToResolve.Assigner.ID) {
		_, err = bot.Send(tgbotapi.NewMessage(int64(taskToResolve.Assigner.ID),
			fmt.Sprintf(`Задача "%s" выполнена @%s`, taskToResolve.Name, chat.UserName),
		))
	}

	return err
}

func resolveIdCommand(ctx context.Context, command string) error {
	parts := strings.Split(command, "_")
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	switch parts[0] {
	case "/assign":
		return assign(ctx, id)
	case "/unassign":
		return unassign(ctx, id)
	case "/resolve":
		return resolve(ctx, id)
	default:
		return errors.New("unknown command: " + command)
	}
}

func createTask(ctx context.Context, taskName string) error {
	chat := ctx.Value("chat").(*tgbotapi.Chat)

	taskStore.mu.Lock()
	id := taskStore.nextID
	taskStore.tasks = append(taskStore.tasks, Task{
		ID:   id,
		Name: taskName,
		Assigner: User{
			ID:       int(chat.ID),
			Username: chat.UserName,
		},
	})
	taskStore.nextID++
	taskStore.mu.Unlock()

	bot := ctx.Value("bot").(*tgbotapi.BotAPI)
	message := tgbotapi.NewMessage(chat.ID, fmt.Sprintf(`Задача "%v" создана, id=%v`, taskName, id))

	_, err := bot.Send(message)

	return err
}

func addAssignee(task Task, chat *tgbotapi.Chat, displayAssignee bool) string {
	if !displayAssignee || task.Assignee.ID == 0 {
		return ""
	}
	if task.Assignee.ID == int(chat.ID) {
		return "\nassignee: я"
	} else {
		return "\nassignee: @" + task.Assignee.Username
	}
}

func addCommands(task Task, chat *tgbotapi.Chat) string {
	if task.Assignee.ID != 0 && task.Assignee.ID != int(chat.ID) {
		return ""
	}

	if task.Assignee.ID != 0 {
		return fmt.Sprintf("\n/unassign_%[1]d /resolve_%[1]d", task.ID)
	} else {
		return "\n/assign_" + strconv.Itoa(int(task.ID))
	}
}

func listTasks(ctx context.Context, tasks []Task, displayAssignee bool) error {
	chat := ctx.Value("chat").(*tgbotapi.Chat)
	bot := ctx.Value("bot").(*tgbotapi.BotAPI)

	if len(tasks) == 0 {
		message := tgbotapi.NewMessage(chat.ID, "Нет задач")

		_, err := bot.Send(message)
		if err != nil {
			return err
		}

		return nil
	}

	var message string
	for idx, task := range tasks {
		message += fmt.Sprintf("%d. %s by @%s", task.ID, task.Name, task.Assigner.Username)
		message += addAssignee(task, chat, displayAssignee) + addCommands(task, chat)

		if idx != (len(tasks) - 1) {
			message += "\n\n"
		}
	}

	tgMessage := tgbotapi.NewMessage(chat.ID, message)
	_, err := bot.Send(tgMessage)

	return err
}

func allTasks(ctx context.Context) error {
	taskStore.mu.RLock()
	tasks := slices.Clone(taskStore.tasks)
	taskStore.mu.RUnlock()

	return listTasks(ctx, tasks, true)
}

func myTasks(ctx context.Context) error {
	defer taskStore.mu.RUnlock()

	chat := ctx.Value("chat").(*tgbotapi.Chat)
	var assignedToMe []Task
	for _, task := range taskStore.tasks {
		if int32(task.Assignee.ID) == int32(chat.ID) {
			assignedToMe = append(assignedToMe, task)
		}
	}
	taskStore.mu.RUnlock()

	return listTasks(ctx, assignedToMe, false)
}

func myOwnedTasks(ctx context.Context) error {
	taskStore.mu.RLock()

	chat := ctx.Value("chat").(*tgbotapi.Chat)
	var ownedByMe []Task
	for _, task := range taskStore.tasks {
		if int64(task.Assigner.ID) == chat.ID {
			ownedByMe = append(ownedByMe, task)
		}
	}
	taskStore.mu.RUnlock()

	return listTasks(ctx, ownedByMe, false)
}

func resolveCommandExecutor(ctx context.Context, command string) error {
	parts := strings.Split(command, " ")
	var err error
	switch parts[0] {
	case "/tasks":
		err = allTasks(ctx)
	case "/new":
		err = createTask(ctx, strings.Join(parts[1:], " "))
	case "/my":
		err = myTasks(ctx)
	case "/owner":
		err = myOwnedTasks(ctx)
	default:
		err = resolveIdCommand(ctx, command)
	}

	return err
}

func startTaskBot(ctx context.Context) error {
	// сюда пишите ваш код
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return err
	}

	go func() {
		http.ListenAndServe(":3000", nil)
	}()

	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		return err
	}
	bot.Request(wh)
	updates := bot.ListenForWebhook("/")
	ctx = context.WithValue(ctx, "bot", bot)

Loop:
	for {
		select {
		case update := <-updates:
			command := update.Message.Text
			ctx = context.WithValue(ctx, "chat", update.Message.Chat)
			err = resolveCommandExecutor(ctx, command)
			if err != nil {
				log.Printf(err.Error())
			}
		case <-ctx.Done():
			break Loop
		}
	}

	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
