package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 640
	screenHeight = 480
	gravity      = 0.5   // 重力の強さ
	damping      = 0.01  // 空気抵抗（0に近づくほど揺れが残る）
	subSteps     = 8     // 1フレーム内の物理計算回数（多いほどロープが硬く、正確になる）
)

type Point struct {
	X, Y         float64
	OldX, OldY   float64
	IsFixed      bool
}

type Link struct {
	P1, P2    *Point
	Length    float64
	Stiffness float64
}

type Game struct {
	points        []*Point
	links         []*Link
	isDragging    bool
	draggedPoint  *Point
}

func NewGame() *Game {
	g := &Game{}

	// 1. 質点の生成（真下に垂れ下がるロープ）
	numPoints := 10
	startY := 50.0
	segmentLength := 30.0

	for i := 0; i < numPoints; i++ {
		p := &Point{
			X:    screenWidth / 2,
			Y:    startY + float64(i)*segmentLength,
			OldX: screenWidth / 2,
			OldY: startY + float64(i)*segmentLength,
		}
		if i == 0 {
			p.IsFixed = true // 一番上の点は固定
		}
		g.points = append(g.points, p)
	}

	// 2. 質点同士をつなぐリンク（バネ）の生成
	for i := 0; i < numPoints-1; i++ {
		l := &Link{
			P1:        g.points[i],
			P2:        g.points[i+1],
			Length:    segmentLength,
			Stiffness: 0.9,
		}
		g.links = append(g.links, l)
	}

	return g
}

func (g *Game) Update() error {
	// --- マウス操作の処理 ---
	mx, my := ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		lastP := g.points[len(g.points)-1]
		dx := float64(mx) - lastP.X
		dy := float64(my) - lastP.Y
		if math.Sqrt(dx*dx+dy*dy) < 20 {
			g.isDragging = true
			g.draggedPoint = lastP
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.isDragging {
		g.draggedPoint.X = float64(mx)
		g.draggedPoint.Y = float64(my)
		g.draggedPoint.OldX = float64(mx)
		g.draggedPoint.OldY = float64(my)
	} else {
		g.isDragging = false
		g.draggedPoint = nil
	}

	// --- 物理演算ループ ---
	for _, p := range g.points {
		if p.IsFixed || p == g.draggedPoint {
			continue
		}
		vx := (p.X - p.OldX) * (1.0 - damping)
		vy := (p.Y - p.OldY) * (1.0 - damping)

		p.OldX = p.X
		p.OldY = p.Y

		p.X += vx
		p.Y += vy + gravity
	}

	for step := 0; step < subSteps; step++ {
		for _, link := range g.links {
			dx := link.P2.X - link.P1.X
			dy := link.P2.Y - link.P1.Y
			distance := math.Sqrt(dx*dx + dy*dy)
			if distance == 0 {
				continue
			}

			difference := link.Length - distance
			percent := (difference / distance) * link.Stiffness * 0.5

			offsetX := dx * percent
			offsetY := dy * percent

			if !link.P1.IsFixed && link.P1 != g.draggedPoint {
				link.P1.X -= offsetX
				link.P1.Y -= offsetY
			}
			if !link.P2.IsFixed && link.P2 != g.draggedPoint {
				link.P2.X += offsetX
				link.P2.Y += offsetY
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 25, 255})

	for _, link := range g.links {
		dx := link.P2.X - link.P1.X
		dy := link.P2.Y - link.P1.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		tensionRatio := (distance - link.Length) / (link.Length * 0.5)
		if tensionRatio < 0 {
			tensionRatio = 0
		}
		if tensionRatio > 1 {
			tensionRatio = 1
		}

		r := uint8(255)
		g := uint8(255 * (1.0 - tensionRatio))
		b := uint8(255 * math.Max(0, 1.0-tensionRatio*2))
		linkColor := color.RGBA{r, g, b, 255}

		vector.StrokeLine(screen, float32(link.P1.X), float32(link.P1.Y), float32(link.P2.X), float32(link.P2.Y), 2, linkColor, true)
	}

	for i, p := range g.points {
		c := color.RGBA{100, 200, 255, 255}
		if p.IsFixed {
			c = color.RGBA{255, 100, 100, 255}
		} else if i == len(g.points)-1 {
			c = color.RGBA{100, 255, 100, 255}
		}
		vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), 4, c, true)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Tension Simulation (Mass-Spring)")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}