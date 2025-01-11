package gui

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Item[T any] struct {
	Label string
	Value T
}

type renderItemFunc[T any] func(index int, item Item[T], selected bool)

type List[T any] struct {
	renderer        *sdl.Renderer
	renderItem      renderItemFunc[T]
	items           []Item[T]
	selectedIndex   int
	scrollOffset    int
	maxVisibleItems int
	position        sdl.Point
}

func NewList[T any](renderer *sdl.Renderer, maxVisibleItems int, position sdl.Point, renderItem renderItemFunc[T]) *List[T] {
	return &List[T]{
		renderer:        renderer,
		renderItem:      renderItem,
		maxVisibleItems: maxVisibleItems,
		items:           []Item[T]{},
		position:        position,
	}
}

func (l *List[T]) SetItems(items []Item[T]) {
	l.items = items
	l.selectedIndex = 0
	l.scrollOffset = 0
}

func (l *List[T]) ScrollDown() {
	if len(l.items) == 0 {
		return
	}
	if l.selectedIndex < len(l.items)-1 {
		l.selectedIndex++
		if l.selectedIndex >= l.scrollOffset+l.maxVisibleItems {
			l.scrollOffset++
		}
	} else {
		l.selectedIndex = len(l.items) - 1
		l.scrollOffset = len(l.items) - l.maxVisibleItems
		if l.scrollOffset < 0 {
			l.scrollOffset = 0
		}
	}
}

func (l *List[T]) ScrollUp() {
	if len(l.items) == 0 {
		return
	}
	if l.selectedIndex > 0 {
		l.selectedIndex--
		if l.selectedIndex < l.scrollOffset {
			l.scrollOffset--
		}
	} else {
		l.selectedIndex = 0
		l.scrollOffset = 0
	}
}

func (l *List[T]) Draw() {
	// Draw the items
	startIndex := l.scrollOffset
	endIndex := startIndex + l.maxVisibleItems
	if endIndex > len(l.items) {
		endIndex = len(l.items)
	}
	visibleItems := l.items[startIndex:endIndex]

	for index, item := range visibleItems {
		selected := index+startIndex == l.selectedIndex
		l.renderItem(index, item, selected)
	}
}

func (l *List[T]) GetSelectedIndex() int {
	return l.selectedIndex
}

func (l *List[T]) GetScrollOffset() int {
	return l.scrollOffset
}

func (l *List[T]) SelectedItem() *T {
	if len(l.items) == 0 {
		var zeroValue T
		return &zeroValue
	}

	return &l.items[l.selectedIndex].Value
}

func (l *List[T]) GetValues() []T {
	values := make([]T, 0, len(l.items))
	for _, item := range l.items {
		values = append(values, item.Value)
	}
	return values
}
