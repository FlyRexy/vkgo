package main

import (
	"sort"
	"strings"
)

/*
	код писать в этом файле
	наверняка у вас будут какие-то структуры с методами, глобальные переменные ( тут можно ), функции
*/

type Room struct {
	name         string
	paths        []string
	items        map[string][]string
	interactWith []string

	lookAroundMessage string
	welcomeMessage    string
	quests            []string
}

type Flat struct {
	rooms  []Room
	locked bool
}

type Player struct {
	currentRoom string
	inventory   []string
}

var p Player
var f Flat

func main() {
	/*
		в этой функции можно ничего не писать,
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
	// a := map[string][]string{
	// 	"a": {"a", "b"},
	// }

	// b := map[string][]int{
	// 	"a": []int{1, 2},
	// }
}

func initGame() {
	/*
		эта функция инициализирует игровой мир - все комнаты
		если что-то было - оно корректно перезатирается
	*/
	p = Player{currentRoom: "кухня"}
	f = Flat{
		rooms: []Room{
			{name: "кухня", welcomeMessage: "кухня, ничего интересного. ", lookAroundMessage: "ты находишься на кухне, ", quests: []string{"собрать рюкзак", "идти в универ"}, paths: []string{"коридор"}, items: map[string][]string{"на столе": {"чай"}}},
			{name: "коридор", welcomeMessage: "ничего интересного. ", paths: []string{"кухня", "комната", "улица"}, interactWith: []string{"дверь"}},
			{
				name:           "комната",
				welcomeMessage: "ты в своей комнате. ",
				paths:          []string{"коридор"},
				items:          map[string][]string{"на столе": {"ключи", "конспекты"}, "на стуле": {"рюкзак"}},
			},
			{name: "улица", welcomeMessage: "на улице весна. ", paths: []string{"домой"}},
		},
		locked: true,
	}
}

func getRoomMeta(roomName string) (*Room, bool) {
	var roomMeta Room
	for idx, room := range f.rooms {
		if room.name == roomName {
			roomMeta = room
			return &f.rooms[idx], true
		}
	}

	return &roomMeta, false
}

func (p Player) lookAround() string {
	roomMeta, _ := getRoomMeta(p.currentRoom)

	placesKeys := []string{}
	for place := range roomMeta.items {
		placesKeys = append(placesKeys, place)
	}
	sort.Strings(placesKeys)

	var itemsMessage string
	for idx, place := range placesKeys {
		if len(roomMeta.items[place]) != 0 {
			itemsMessage += place + ": " + strings.Join(roomMeta.items[place], ", ")
		}
		if idx != len(placesKeys)-1 && itemsMessage != "" {
			itemsMessage += ", "
		}
	}
	if itemsMessage == "" {
		itemsMessage += "пустая комната"
	}

	pathMessage := ". можно пройти - " + strings.Join(roomMeta.paths, ", ")
	var questsMessage string
	if len(roomMeta.quests) != 0 {
		questsMessage = ", надо " + strings.Join(roomMeta.quests, " и ")
	}

	return roomMeta.lookAroundMessage + itemsMessage + questsMessage + pathMessage
}

func (p *Player) goTo(nextRoom string) string {
	roomMeta, _ := getRoomMeta(nextRoom)
	currentRoomMeta, _ := getRoomMeta(p.currentRoom)
	var found bool
	for _, path := range currentRoomMeta.paths {
		if path == nextRoom {
			found = true
			break
		}
	}
	if !found {
		return "нет пути в " + nextRoom
	}
	if nextRoom == "улица" && f.locked {
		return "дверь закрыта"
	}

	p.currentRoom = nextRoom

	pathMessage := "можно пройти - " + strings.Join(roomMeta.paths, ", ")
	return roomMeta.welcomeMessage + pathMessage
}

func (p *Player) take(thing string) string {
	var isHere bool
	roomMeta, _ := getRoomMeta(p.currentRoom)
	for _, placeItems := range roomMeta.items {
		for _, item := range placeItems {
			if item == thing {
				isHere = true
				break
			}
		}
	}
	if !isHere {
		return "нет такого"
	}
	hasBackpack := false
	for _, item := range p.inventory {
		if item == "рюкзак" {
			hasBackpack = true
			break
		}
	}

	if !hasBackpack {
		return "некуда класть"
	}

	for place, placeItems := range roomMeta.items {
		newItems := placeItems[:0]

		for _, item := range placeItems {
			if item != thing {
				newItems = append(newItems, item)
			}
		}

		if len(newItems) != 0 {
			roomMeta.items[place] = newItems
		} else {
			delete(roomMeta.items, place)
		}
	}

	p.inventory = append(p.inventory, thing)

	return "предмет добавлен в инвентарь: " + thing
}

func (p *Player) apply(item, to string) string {
	hasItem := false
	for _, playerItem := range p.inventory {
		if playerItem == item {
			hasItem = true
			break
		}
	}
	if !hasItem {
		return "нет предмета в инвентаре - " + item
	}

	roomMeta, _ := getRoomMeta(p.currentRoom)
	hasInteraction := false
	for _, interaction := range roomMeta.interactWith {
		if interaction == to {
			hasInteraction = true
			break
		}
	}
	if !hasInteraction {
		return "не к чему применить"
	}

	f.locked = false
	return "дверь открыта"
}

func (p *Player) putOn(thing string) string {
	if thing != "рюкзак" {
		return ""
	}
	roomMeta, _ := getRoomMeta(p.currentRoom)

	for place, placeItems := range roomMeta.items {
		newItems := placeItems[:0]

		for _, item := range placeItems {
			if item != thing {
				newItems = append(newItems, item)
			}
		}

		if len(newItems) != 0 {
			roomMeta.items[place] = newItems
		} else {
			delete(roomMeta.items, place)
		}
	}
	p.inventory = append(p.inventory, thing)
	for idx, room := range f.rooms {
		if room.name == "кухня" {
			updatedQuests := f.rooms[idx].quests[:0]

			for _, quest := range f.rooms[idx].quests {
				if quest != "собрать рюкзак" {
					updatedQuests = append(updatedQuests, quest)
				}
			}

			f.rooms[idx].quests = updatedQuests
		}
	}
	return "вы надели: рюкзак"

}

func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/
	parameters := strings.Split(command, " ")
	switch parameters[0] {
	case "осмотреться":
		return p.lookAround()
	case "идти":
		return p.goTo(parameters[1])
	case "взять":
		return p.take(parameters[1])
	case "надеть":
		return p.putOn(parameters[1])
	case "применить":
		return p.apply(parameters[1], parameters[2])
	default:
		return "неизвестная команда"
	}
}
