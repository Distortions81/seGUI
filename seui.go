package seGUI

import (
	"errors"
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	titleBarSize        = 24
	halfTitleBarSize    = titleBarSize / 2
	quarterTitleBarSize = titleBarSize / 4
)

var (
	windowList  map[string]*windowObject
	openWindows []*windowObject

	windowsLock sync.Mutex

	whiteImage    = ebiten.NewImage(3, 3)
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	windowList = map[string]*windowObject{}
	whiteImage.Fill(color.White)
}

// Run this in ebiten draw(), pass "screen"
func DrawWindows(screen *ebiten.Image) {

	windowsLock.Lock()
	defer windowsLock.Unlock()

	for _, win := range openWindows {
		//Draw window
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(win.position.X), float64(win.position.Y))
		screen.DrawImage(win.drawCache, op)

		if !win.win.HasTitleBar {
			continue
		}

		vector.DrawFilledRect(screen, float32(win.position.X), float32(win.position.Y),
			float32(win.size.X), titleBarSize, win.win.TitleBGColor, false)

		tr := V2i{X: win.position.X + win.size.X - quarterTitleBarSize, Y: win.position.Y + quarterTitleBarSize}

		//Draw close X
		if win.win.Closable {
			var path vector.Path
			path.MoveTo(float32(tr.X), float32(tr.Y))
			path.LineTo(float32(tr.X-halfTitleBarSize), float32(tr.Y+halfTitleBarSize))

			path.MoveTo(float32(tr.X-halfTitleBarSize), float32(tr.Y))
			path.LineTo(float32(tr.X), float32(tr.Y+halfTitleBarSize))

			path.Close()

			var vs []ebiten.Vertex
			var is []uint16
			vop := &vector.StrokeOptions{Width: 5, LineJoin: vector.LineJoinRound, LineCap: vector.LineCapRound}
			vs, is = path.AppendVerticesAndIndicesForStroke(nil, nil, vop)

			red, green, blue, alpha := win.win.TitleButtonColor.RGBA()
			for i := range vs {
				vs[i].ColorR = float32(red / 255)
				vs[i].ColorG = float32(green / 255)
				vs[i].ColorB = float32(blue / 255)
				vs[i].ColorA = float32(alpha / 255)
			}

			top := &ebiten.DrawTrianglesOptions{AntiAlias: true, FillRule: ebiten.FillAll}
			screen.DrawTriangles(vs, is, whiteSubImage, top)
		}
	}
}

// Add a window. Returns true if added
func AddWindow(windowID string, window WindowData) error {
	windowsLock.Lock()
	defer windowsLock.Unlock()

	newWin := &windowObject{win: window, dirty: true}

	newWin.size = newWin.win.StartSize
	if window.HasTitleBar {
		newWin.size.Y += titleBarSize
	}
	newWin.position = newWin.win.StartPosition
	windowList[windowID] = newWin

	newWin.drawCache = ebiten.NewImage(newWin.size.X, newWin.size.Y)
	if newWin.drawCache == nil {
		return errors.New("unable to create window draw cache")
	}

	newWin.drawCache.Fill(newWin.win.BGColor)
	return nil
}

// Delete a window. Returns true if deleted
func DeleteWindow(windowID string) error {
	windowsLock.Lock()
	defer windowsLock.Unlock()

	if windowList[windowID] != nil {
		delete(windowList, windowID)
		return nil
	}

	return errors.New("unable to find window")
}

// Update a window. Returns true if updated
func UpdateWindow(windowID string, window WindowData) error {
	windowsLock.Lock()
	defer windowsLock.Unlock()

	if windowList[windowID] != nil {
		windowList[windowID].win = window
		windowList[windowID].dirty = true
		return nil
	}

	return errors.New("unable to find window")
}

// Update window items. Returns true if updated
func UpdateWindowItems(windowID string, windowItems []WindowItemData) error {
	windowsLock.Lock()
	defer windowsLock.Unlock()

	if windowList[windowID] != nil {
		windowList[windowID].items = windowItems
		windowList[windowID].dirty = true
		return nil
	}

	return errors.New("unable to find window")
}

func OpenWindow(windowID string) error {
	windowsLock.Lock()
	defer windowsLock.Unlock()

	window := windowList[windowID]

	if window != nil {
		if !window.open {
			window.open = true

			for _, win := range openWindows {
				if win.id == windowID {
					return nil
				}
			}
			openWindows = append(openWindows, window)
		}
		return nil
	}

	return errors.New("unable to find window")
}

func CloseWindow(windowID string) error {
	windowsLock.Lock()
	defer windowsLock.Unlock()

	window := windowList[windowID]

	if window != nil {
		if window.open {
			window.open = false

			numOpen := len(openWindows) - 1
			for w := numOpen; numOpen > 0; numOpen-- {
				if openWindows[w].id != windowID {

					//Delete item
					openWindows = append(openWindows[:w], openWindows[w+1:]...)
				}
			}
		}
		return nil
	}

	return errors.New("unable to find window")
}
