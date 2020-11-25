package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/font/opentype"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Level struct {
	mazeWall [10]Wall
	maxWall  int
	level    int
}

type Wall struct {
	xLoc int
	yLoc int
	pict *ebiten.Image
}

type Weapon struct {
	pict      *ebiten.Image
	dx        int
	dy        int
	direction string
	enemyShot bool
}

type Sprite struct {
	pict       *ebiten.Image
	xLoc       int
	yLoc       int
	dx         int
	dy         int
	Weapon     Weapon
	activeShot bool
	lives      int
	alive      bool
	hitWall    bool
}

type InfoBar struct {
	imageBar   *ebiten.Image
	playerName string
	score      int
	playerNum  int
}

type Game struct {
	playerSprite Sprite
	khaiSprite   [numEnemies]Sprite
	sophiaSprite [numEnemies]Sprite
	drawOps      ebiten.DrawImageOptions
	infoBar      InfoBar
	wall         [4]Wall
	outOfBounds  bool
	level        [6]Level
	currentLevel int
	counter      int
}

const (
	GameTitle        = "Run Pupperooo~!!"
	GameInstructions = "You are a cute dog ready to take your tenth nap of the day but you are surrounded by playful toddlers. " +
		"These toddlers \nare chasing you down because want to play with you and this is getting in the way of your nap. " +
		"You must throw \nfrisbees at them to make them go away otherwise they will shoot at you with their squirt guns and you HATE squirt guns.\n" +
		"\n** Instructions **\n" +
		"Your frisbee has magical powers and make toddlers fall asleep without hurting them. Hit all the toddlers and move \nonto the next level. " +
		"Once all the levels are completed, you are finally able to take your nap. " +
		"You will have three \nlives. " +
		"and if you get shot with a squirt gun or run into the wall, you will lose a live \n" +
		"\n** Controls ** - on the keyboard press the following\n" +
		"        'W' - SHOOT UP\n        'S' - SHOOT DOWN\n        'A' - SHOOT RIGHT\n        'D' - SHOOT RIGHT\n" +
		"        'UP ARROW' - MOVE UP\n        'DOWN ARROW' - MOVE DOWN\n        'RIGHT ARROW' - MOVE RIGHT\n        'LEFT ARROW' - MOVE LEFT"
	ScreenWidth       = 1000
	ScreenHeight      = 750
	WallThickness     = 10
	InfoBarHeight     = 40
	TotalScreenHeight = ScreenHeight + InfoBarHeight
	numEnemies        = 3
	khaiValue         = 200
	sophiaValue       = 500
)

var (
	xStart        = 30
	yStart        = 70
	xKhai         int
	yKhai         int
	xSophia       int
	ySophia       int
	deadSprite    = -9999
	playerText    string
	textColor     color.RGBA
	updateScore   = 0
	randDirection int
)

func resetPlayer(game *Game) {
	game.playerSprite.xLoc = xStart
	game.playerSprite.yLoc = yStart
	game.infoBar.score -= 100
	game.playerSprite.lives--
}

func outOfBounds(picture *ebiten.Image, xLoc int, yLoc int) bool {
	pictureWidth, pictureHeight := picture.Size()
	if xLoc <= WallThickness ||
		xLoc+pictureWidth >= ScreenWidth-WallThickness ||
		yLoc <= WallThickness+InfoBarHeight ||
		yLoc+pictureHeight >= ScreenHeight-WallThickness {
		return true
	}
	return false
}

func weaponOrEnemyOut(game *Game) {
	for i := 0; i < numEnemies; i++ {
		if game.khaiSprite[i].alive == false {
			game.khaiSprite[i].xLoc = deadSprite
			game.khaiSprite[i].yLoc = deadSprite
		}
		if game.sophiaSprite[i].alive == false {
			game.sophiaSprite[i].xLoc = deadSprite
			game.sophiaSprite[i].yLoc = deadSprite
		}
	}
	if game.playerSprite.activeShot == false {
		game.playerSprite.Weapon.dy = -3000
		game.playerSprite.Weapon.dx = -3000
	}
}

func endMovement(game *Game) {
	speed := 3
	spriteW, _ := game.khaiSprite[0].pict.Size()
	game.sophiaSprite[0].xLoc += speed
	game.khaiSprite[0].xLoc += speed
	game.playerSprite.xLoc += speed

	if game.sophiaSprite[0].xLoc-spriteW > ScreenWidth {
		game.sophiaSprite[0].xLoc = 0
	}
	if game.khaiSprite[0].xLoc-spriteW > ScreenWidth {
		game.khaiSprite[0].xLoc = 0
	}
	if game.playerSprite.xLoc-spriteW > ScreenWidth {
		game.playerSprite.xLoc = 0
	}
}

func hitMaze(game *Game) {
	ammoWidth, ammoHeight := game.playerSprite.Weapon.pict.Size()
	playerWidth, playerHeight := game.playerSprite.pict.Size()

	for i := 0; i < game.level[game.currentLevel].maxWall; i++ {
		wallWidth, wallHeight := game.level[game.currentLevel].mazeWall[i].pict.Size()
		xPict, yPict := getWallLocation(game.level[game.currentLevel].mazeWall[i])

		// if player or player's ammo hits maze wall - disappear or lose a life
		if game.playerSprite.Weapon.dx > xPict && game.playerSprite.Weapon.dx < xPict+wallWidth &&
			game.playerSprite.Weapon.dy > yPict && game.playerSprite.Weapon.dy < yPict+wallHeight ||
			game.playerSprite.Weapon.dx+ammoWidth > xPict && game.playerSprite.Weapon.dx+ammoWidth < xPict+wallWidth &&
				game.playerSprite.Weapon.dy+ammoHeight > yPict && game.playerSprite.Weapon.dy+ammoHeight < yPict+wallHeight {
			game.playerSprite.activeShot = false
		}
		if game.playerSprite.xLoc > xPict && game.playerSprite.xLoc < xPict+wallWidth &&
			game.playerSprite.yLoc > yPict && game.playerSprite.yLoc < yPict+wallHeight ||
			game.playerSprite.xLoc+playerWidth > xPict && game.playerSprite.xLoc+playerWidth < xPict+wallWidth &&
				game.playerSprite.yLoc+playerHeight > yPict && game.playerSprite.yLoc+playerHeight < yPict+wallHeight {
			resetPlayer(game)
		}

		enemyWidth, enemyHeight := game.khaiSprite[0].pict.Size()
		for i := 0; i < numEnemies; i++ {
			// if enemies hit maze wall - disappear
			if game.sophiaSprite[i].xLoc > xPict && game.sophiaSprite[i].xLoc < xPict+wallWidth &&
				game.sophiaSprite[i].yLoc > yPict && game.sophiaSprite[i].yLoc < yPict+wallHeight ||
				game.sophiaSprite[i].xLoc+enemyWidth > xPict && game.sophiaSprite[i].xLoc+enemyWidth < xPict+wallWidth &&
					game.sophiaSprite[i].yLoc+enemyHeight > yPict && game.sophiaSprite[i].yLoc+enemyHeight < yPict+wallHeight {
				game.sophiaSprite[i].alive = false
				game.infoBar.score += (sophiaValue / 2)
			}
			if game.khaiSprite[i].xLoc > xPict && game.khaiSprite[i].xLoc < xPict+wallWidth &&
				game.khaiSprite[i].yLoc > yPict && game.khaiSprite[i].yLoc < yPict+wallHeight ||
				game.khaiSprite[i].xLoc+enemyWidth > xPict && game.khaiSprite[i].xLoc+enemyWidth < xPict+wallWidth &&
					game.khaiSprite[i].yLoc+enemyHeight > yPict && game.khaiSprite[i].yLoc+enemyHeight < yPict+wallHeight {
				game.khaiSprite[i].alive = false
				game.infoBar.score += (khaiValue / 2)
			}

			// if enemy shots hit maze wall - disappear
			if game.sophiaSprite[i].Weapon.dx > xPict && game.sophiaSprite[i].Weapon.dx < xPict+wallWidth &&
				game.sophiaSprite[i].Weapon.dy > yPict && game.sophiaSprite[i].Weapon.dy < yPict+wallHeight ||
				game.sophiaSprite[i].Weapon.dx+ammoWidth > xPict && game.sophiaSprite[i].Weapon.dx+ammoWidth < xPict+wallWidth &&
					game.sophiaSprite[i].Weapon.dy+ammoHeight > yPict && game.sophiaSprite[i].dy+ammoHeight < yPict+wallHeight {
				game.sophiaSprite[i].activeShot = false
			}

			// player collision with enemy sprites
			if game.playerSprite.xLoc > game.khaiSprite[i].xLoc && game.playerSprite.xLoc < game.khaiSprite[i].xLoc+enemyHeight &&
				game.playerSprite.yLoc > game.khaiSprite[i].yLoc && (game.playerSprite.yLoc) < game.khaiSprite[i].yLoc+enemyWidth ||
				game.playerSprite.xLoc+playerWidth > game.khaiSprite[i].xLoc && game.playerSprite.xLoc+playerWidth < game.khaiSprite[i].xLoc+enemyHeight &&
					game.playerSprite.yLoc+playerHeight > game.khaiSprite[i].yLoc && game.playerSprite.yLoc+playerHeight < game.khaiSprite[i].yLoc+enemyWidth {
				resetPlayer(game)
			}
			if game.playerSprite.xLoc > game.sophiaSprite[i].xLoc && game.playerSprite.xLoc < game.sophiaSprite[i].xLoc+enemyHeight &&
				game.playerSprite.yLoc > game.sophiaSprite[i].yLoc && (game.playerSprite.yLoc) < game.sophiaSprite[i].yLoc+enemyWidth ||
				game.playerSprite.xLoc+playerWidth > game.sophiaSprite[i].xLoc && game.playerSprite.xLoc+playerWidth < game.sophiaSprite[i].xLoc+enemyHeight &&
					game.playerSprite.yLoc+playerHeight > game.sophiaSprite[i].yLoc && game.playerSprite.yLoc+playerHeight < game.sophiaSprite[i].yLoc+enemyWidth {
				resetPlayer(game)
			}
		}
	}
}

func enemyMovement(enemy Sprite, game *Game) Sprite {
	movementSpeed := 10
	if game.counter%200 == 0 {
		if enemy.xLoc < game.playerSprite.xLoc {
			enemy.xLoc += movementSpeed
		} else {
			enemy.xLoc -= movementSpeed
		}
		if enemy.yLoc < game.playerSprite.yLoc {
			enemy.yLoc += movementSpeed
		} else {
			enemy.yLoc -= movementSpeed
		}
	}
	return enemy
}

func playerMovement(game *Game) {
	playerspeed := 5
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		game.playerSprite.dx = -playerspeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		game.playerSprite.dx = playerspeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyLeft) || inpututil.IsKeyJustReleased(ebiten.KeyRight) {
		game.playerSprite.dx = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		game.playerSprite.dy = -playerspeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		game.playerSprite.dy = playerspeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyUp) || inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		game.playerSprite.dy = 0
	}
	game.playerSprite.yLoc += game.playerSprite.dy
	game.playerSprite.xLoc += game.playerSprite.dx
}

func isShooting(game *Game) {
	ammoHeight, ammoWidth := game.playerSprite.Weapon.pict.Size()

	if inpututil.IsKeyJustPressed(ebiten.KeyD) ||
		inpututil.IsKeyJustPressed(ebiten.KeyA) ||
		inpututil.IsKeyJustPressed(ebiten.KeyS) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) {

		game.playerSprite.Weapon.dx = game.playerSprite.xLoc + (ammoWidth)
		game.playerSprite.Weapon.dy = game.playerSprite.yLoc + (ammoHeight)
		game.playerSprite.activeShot = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		game.playerSprite.Weapon.direction = "right"
	} else if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		game.playerSprite.Weapon.direction = "left"
	} else if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		game.playerSprite.Weapon.direction = "down"
	} else if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		game.playerSprite.Weapon.direction = "up"
	}
}

func enemyShooting(enemy Sprite, game *Game) Sprite {
	ammoHeight, ammoWidth := game.playerSprite.Weapon.pict.Size()

	if game.counter%100 == 0 {
		randDirection = rand.Intn(4)
		enemy.Weapon.dx = enemy.xLoc + (ammoWidth)
		enemy.Weapon.dy = enemy.yLoc + (ammoHeight)
		enemy.activeShot = true
	}
	if randDirection == 0 {
		enemy.Weapon.direction = "up"
	} else if randDirection == 1 {
		enemy.Weapon.direction = "down"
	} else if randDirection == 2 {
		enemy.Weapon.direction = "left"
	} else {
		enemy.Weapon.direction = "right"
	}
	return enemy
}

func enemyHit(enemy Sprite, game *Game) {
	enemyH, enemyW := enemy.pict.Size()
	ammoH, ammoW := game.playerSprite.Weapon.pict.Size()
	if game.playerSprite.Weapon.dx > enemy.xLoc && game.playerSprite.Weapon.dx < enemy.xLoc+enemyH &&
		game.playerSprite.Weapon.dy > enemy.yLoc && game.playerSprite.Weapon.dy < enemy.yLoc+enemyW ||
		game.playerSprite.Weapon.dx+ammoH > enemy.xLoc && game.playerSprite.Weapon.dx+ammoH < enemy.xLoc+enemyH &&
			game.playerSprite.Weapon.dy+ammoW > enemy.yLoc && game.playerSprite.Weapon.dy+ammoW < enemy.yLoc+enemyW {
		game.playerSprite.Weapon.enemyShot = true
	}
}

func (game *Game) Update() error {
	game.counter++

	if game.currentLevel == 0 || game.currentLevel == 4 {
		for i := 1; i < numEnemies; i++ {
			game.khaiSprite[i].xLoc = deadSprite
			game.khaiSprite[i].yLoc = deadSprite
			game.sophiaSprite[i].xLoc = deadSprite
			game.sophiaSprite[i].yLoc = deadSprite
		}
	}

	if game.currentLevel != 0 && game.currentLevel != 4 {
		weaponOrEnemyOut(game)
		playerMovement(game)
		game.outOfBounds = outOfBounds(game.playerSprite.pict, game.playerSprite.xLoc, game.playerSprite.yLoc)
		if game.outOfBounds == true {
			resetPlayer(game)
		}

		for i := 0; i < numEnemies; i++ {
			game.sophiaSprite[i] = enemyShooting(game.sophiaSprite[i], game)
			if game.sophiaSprite[i].activeShot {
				shootSpeed := 8
				if game.sophiaSprite[i].Weapon.direction == "right" {
					game.sophiaSprite[i].Weapon.dx += shootSpeed
				} else if game.sophiaSprite[i].Weapon.direction == "left" {
					game.sophiaSprite[i].Weapon.dx -= shootSpeed
				} else if game.sophiaSprite[i].Weapon.direction == "up" {
					game.sophiaSprite[i].Weapon.dy -= shootSpeed
				} else if game.sophiaSprite[i].Weapon.direction == "down" {
					game.sophiaSprite[i].Weapon.dy += shootSpeed
				}
				if outOfBounds(game.sophiaSprite[i].Weapon.pict, game.sophiaSprite[i].Weapon.dx, game.sophiaSprite[i].Weapon.dy) {
					game.sophiaSprite[i].activeShot = false
				}
			}
		}

		isShooting(game)
		if game.playerSprite.activeShot {
			shootSpeed := 8
			if game.playerSprite.Weapon.direction == "right" {
				game.playerSprite.Weapon.dx += shootSpeed
			} else if game.playerSprite.Weapon.direction == "left" {
				game.playerSprite.Weapon.dx -= shootSpeed
			} else if game.playerSprite.Weapon.direction == "up" {
				game.playerSprite.Weapon.dy -= shootSpeed
			} else if game.playerSprite.Weapon.direction == "down" {
				game.playerSprite.Weapon.dy += shootSpeed
			}
			if outOfBounds(game.playerSprite.Weapon.pict, game.playerSprite.Weapon.dx, game.playerSprite.Weapon.dy) {
				game.playerSprite.activeShot = false
			}

			for i := 0; i < numEnemies; i++ {
				enemyHit(game.khaiSprite[i], game)
				if game.playerSprite.Weapon.enemyShot {
					game.khaiSprite[i].lives--
					game.playerSprite.Weapon.enemyShot = false
					game.playerSprite.activeShot = false
					game.infoBar.score += khaiValue
					if game.khaiSprite[i].lives <= 0 {
						game.khaiSprite[i].alive = false
						game.infoBar.score += 300
					}
				}
				enemyHit(game.sophiaSprite[i], game)
				if game.playerSprite.Weapon.enemyShot {
					game.sophiaSprite[i].lives--
					game.playerSprite.Weapon.enemyShot = false
					game.playerSprite.activeShot = false
					game.infoBar.score += sophiaValue
					if game.sophiaSprite[i].lives <= 0 {
						game.sophiaSprite[i].alive = false
					}
				}

				if outOfBounds(game.khaiSprite[i].Weapon.pict, game.khaiSprite[i].Weapon.dx, game.khaiSprite[i].Weapon.dy) {
					game.khaiSprite[i].activeShot = false
				}
			}
		} // end of shot handler if statement

		for i := 0; i < numEnemies; i++ {
			game.sophiaSprite[i] = enemyMovement(game.sophiaSprite[i], game)
			game.khaiSprite[i] = enemyMovement(game.khaiSprite[i], game)
		}
		hitMaze(game)

		// if you beat a level
		for i := 0; i < numEnemies; i++ {
			if game.sophiaSprite[0].alive == false && game.sophiaSprite[1].alive == false && game.sophiaSprite[2].alive == false &&
				game.khaiSprite[0].alive == false && game.khaiSprite[1].alive == false && game.khaiSprite[2].alive == false {
				game.playerSprite.xLoc = xStart
				game.playerSprite.yLoc = yStart
				setEnemyLocation(game)
				if game.currentLevel == 3 { // set positions for end screen effect if game was beat
					game.khaiSprite[0].xLoc = 50
					game.khaiSprite[0].yLoc = ScreenHeight - 100
					game.sophiaSprite[0].xLoc = 150
					game.sophiaSprite[0].yLoc = ScreenHeight - 100
					game.playerSprite.xLoc = 275
					game.playerSprite.yLoc = ScreenHeight - 100
				}
				game.currentLevel++

				for i := 0; i < numEnemies; i++ {
					game.khaiSprite[i].alive = true
					game.sophiaSprite[i].alive = true
				}
			}
		}

		if game.playerSprite.lives < 0 { // if you died
			game.khaiSprite[0].xLoc = 50
			game.khaiSprite[0].yLoc = ScreenHeight - 100
			game.sophiaSprite[0].xLoc = 150
			game.sophiaSprite[0].yLoc = ScreenHeight - 100
			game.playerSprite.xLoc = 275
			game.playerSprite.yLoc = ScreenHeight - 100
			UpdateScore(game.infoBar.score, game.infoBar.playerNum)
			updateScore++
			game.currentLevel = 4
		}

	} else if game.currentLevel == 4 {
		if updateScore == 0 && game.playerSprite.lives >= 0 {
			UpdateScore(game.infoBar.score, game.infoBar.playerNum)
			updateScore++
		}
		endMovement(game)

	} else { // if current level is 0 - start game window
		game.khaiSprite[0].xLoc = 250
		game.khaiSprite[0].yLoc = 100
		game.playerSprite.xLoc = 400
		game.playerSprite.yLoc = 100
		game.sophiaSprite[0].xLoc = 700
		game.sophiaSprite[0].yLoc = 100

		textColor.R = 0x80 + uint8(rand.Intn(0x7f))
		textColor.G = 0x80 + uint8(rand.Intn(0x7f))
		textColor.B = 0x80 + uint8(rand.Intn(0x7f))
		textColor.A = 0xff

		game.infoBar.playerName += string(ebiten.InputChars())
		if playerTyping(ebiten.KeyEnter) && len(playerText) > 0 || playerTyping(ebiten.KeyKPEnter) && len(playerText) > 0 {
			game.infoBar.playerName = playerText
			AddPlayerName(game.infoBar.playerName)
			game.infoBar.playerNum = GetPlayerNum()
			setEnemyLocation(game)
			game.currentLevel++
		}
		if playerTyping(ebiten.KeyBackspace) {
			if len(game.infoBar.playerName) >= 1 {
				game.infoBar.playerName = game.infoBar.playerName[:len(game.infoBar.playerName)-1]
			}
		}
		if len(game.infoBar.playerName) >= 16 {
			game.infoBar.playerName = game.infoBar.playerName[:len(game.infoBar.playerName)-1]
		}

	}
	return nil
} // end of Update

func setEnemyLocation(game *Game) {
	for i := 0; i < numEnemies; i++ {
		enemyWidth, enemyHeight := game.khaiSprite[0].pict.Size()
		min := 50
		maxHeight := ScreenHeight - enemyHeight - WallThickness
		maxWidth := ScreenWidth - enemyWidth - WallThickness
		xKhai = rand.Intn(maxWidth-min) + min
		yKhai = rand.Intn(maxHeight-min) + min
		game.khaiSprite[i].xLoc = xKhai
		game.khaiSprite[i].yLoc = yKhai
		xSophia = rand.Intn(maxWidth-min) + min
		ySophia = rand.Intn(maxHeight-min) + min
		game.sophiaSprite[i].xLoc = xSophia
		game.sophiaSprite[i].yLoc = ySophia
	}
}

func playerTyping(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	duration := inpututil.KeyPressDuration(key)
	if duration == 1 {
		return true
	}
	if duration >= delay && (duration-delay)%interval == 0 {
		return true
	}
	return false
}

// ----------------------------------------------------- End of Update and its Functions --------------------------------

func (game Game) DrawEnemySprites(screen *ebiten.Image) {
	for i := 0; i < numEnemies; i++ {
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.khaiSprite[i].xLoc), float64(game.khaiSprite[i].yLoc))
		screen.DrawImage(game.khaiSprite[i].pict, &game.drawOps)

		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.sophiaSprite[i].xLoc), float64(game.sophiaSprite[i].yLoc))
		screen.DrawImage(game.sophiaSprite[i].pict, &game.drawOps)
	}
}

func (game Game) DrawPlayerSprite(screen *ebiten.Image) {
	game.drawOps.GeoM.Reset()
	game.drawOps.GeoM.Translate(float64(game.playerSprite.xLoc), float64(game.playerSprite.yLoc))
	screen.DrawImage(game.playerSprite.pict, &game.drawOps)
}

func (game Game) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Moccasin)

	if game.currentLevel != 0 && game.currentLevel != 4 {
		// draw player
		game.DrawPlayerSprite(screen)

		// draw shots
		if game.playerSprite.activeShot == true {
			game.drawOps.GeoM.Reset()
			game.drawOps.GeoM.Translate(float64(game.playerSprite.Weapon.dx), float64(game.playerSprite.Weapon.dy))
			screen.DrawImage(game.playerSprite.Weapon.pict, &game.drawOps)
		}
		for i := 0; i < numEnemies; i++ {
			if game.khaiSprite[i].activeShot == true {
				game.drawOps.GeoM.Reset()
				game.drawOps.GeoM.Translate(float64(game.khaiSprite[i].Weapon.dx), float64(game.khaiSprite[i].Weapon.dy))
				screen.DrawImage(game.khaiSprite[i].Weapon.pict, &game.drawOps)
			}
			if game.sophiaSprite[i].activeShot == true {
				game.drawOps.GeoM.Reset()
				game.drawOps.GeoM.Translate(float64(game.sophiaSprite[i].Weapon.dx), float64(game.sophiaSprite[i].Weapon.dy))
				screen.DrawImage(game.sophiaSprite[i].Weapon.pict, &game.drawOps)
			}
		}
		game.drawWall(screen, game.currentLevel)

		// draw enemy
		for i := 0; i < numEnemies; i++ {
			if game.khaiSprite[i].alive == true {
				game.drawOps.GeoM.Reset()
				game.drawOps.GeoM.Translate(float64(game.khaiSprite[i].xLoc), float64(game.khaiSprite[i].yLoc))
				screen.DrawImage(game.khaiSprite[i].pict, &game.drawOps)
			}
			if game.sophiaSprite[i].alive == true {
				game.drawOps.GeoM.Reset()
				game.drawOps.GeoM.Translate(float64(game.sophiaSprite[i].xLoc), float64(game.sophiaSprite[i].yLoc))
				screen.DrawImage(game.sophiaSprite[i].pict, &game.drawOps)
			}
		}
		game.GameInfoBar(screen)

	} else if game.currentLevel == 4 { // end game
		TopFive, LastHighScore = GetTopFive(SortedPlayers)

		endBar := ebiten.NewImage(ScreenWidth, ScreenHeight-150)
		endBar.Fill(colornames.Black)
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(0, 0)
		screen.DrawImage(endBar, &game.drawOps)
		text.Draw(screen, "Game Over!", makeFont(64, 72), 300, 100, colornames.Tomato)
		text.Draw(screen, game.infoBar.playerName, makeFont(30, 72), 200, 200, colornames.White)
		text.Draw(screen, "score: "+strconv.Itoa(game.infoBar.score), makeFont(30, 72), 200, 250, colornames.White)
		if game.playerSprite.lives < 0 {
			text.Draw(screen, "You Lost!", makeFont(30, 72), 250, 300, colornames.White)
		} else {
			text.Draw(screen, "You Won!", makeFont(30, 72), 250, 300, colornames.White)
		}
		text.Draw(screen, "Current High Scores", makeFont(30, 72), 600, 200, colornames.White)
		for i := range TopFive {
			yAxis := 50 * i
			text.Draw(screen, TopFive[i], makeFont(25, 72), 600, 250+yAxis, colornames.White)
		}
		if game.infoBar.score >= LastHighScore {
			text.Draw(screen, "A new high Score!!", makeFont(30, 72), 200, 350, colornames.White)
		}

		game.DrawPlayerSprite(screen)
		game.DrawEnemySprites(screen)

	} else { // start game
		gameBar := ebiten.NewImage(ScreenWidth-20, ScreenHeight-220)
		gameBar.Fill(colornames.Black)
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(10, 210)
		screen.DrawImage(gameBar, &game.drawOps)

		playerText = game.infoBar.playerName
		text.Draw(screen, "Welcome to "+GameTitle, makeFont(48, 72), 150, 280, textColor)
		text.Draw(screen, GameInstructions, makeFont(14, 72), 50, 320, color.White)
		text.Draw(screen, "Enter your name: "+game.infoBar.playerName, makeFont(20, 72), ScreenWidth-400, ScreenHeight-100, color.White)

		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.playerSprite.xLoc), float64(game.playerSprite.yLoc))
		screen.DrawImage(game.playerSprite.pict, &game.drawOps)

		game.DrawPlayerSprite(screen)
		game.DrawEnemySprites(screen)
	}

	tps := fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS())
	fps := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	text.Draw(screen, tps+"\n"+fps, makeFont(8, 72), 950, 10, color.White)

} // end of draw

func makeFont(size int, dpi int) font.Face {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	theFont, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(size),
		DPI:     float64(dpi),
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	return theFont
}

func (game *Game) GameInfoBar(screen *ebiten.Image) {

	game.infoBar.playerName = playerText

	infoBar := ebiten.NewImage(ScreenWidth, InfoBarHeight)
	infoBar.Fill(colornames.Black)
	game.infoBar.imageBar = infoBar
	gameFont := font.Face(inconsolata.Regular8x16)
	text.Draw(infoBar, "Player Name: "+game.infoBar.playerName, gameFont, 20, 25, color.White)
	text.Draw(infoBar, "#: "+strconv.Itoa(game.infoBar.playerNum), gameFont, 300, 25, color.White)
	text.Draw(infoBar, "Score: "+strconv.Itoa(game.infoBar.score), gameFont, 450, 25, color.White)
	text.Draw(infoBar, "Lives: "+strconv.Itoa(game.playerSprite.lives), gameFont, 720, 25, color.White)
	text.Draw(infoBar, "Level: "+strconv.Itoa(game.currentLevel), gameFont, 850, 25, color.White)

	game.drawOps.GeoM.Reset()
	game.drawOps.GeoM.Translate(0, 0)
	screen.DrawImage(game.infoBar.imageBar, &game.drawOps)
}

func (game Game) drawWall(screen *ebiten.Image, level int) {
	// surrounding walls
	game.drawOps.GeoM.Reset()
	game.drawOps.GeoM.Translate(0, WallThickness*4)
	screen.DrawImage(game.wall[0].pict, &game.drawOps)
	game.drawOps.GeoM.Reset()
	screen.DrawImage(game.wall[2].pict, &game.drawOps)
	game.drawOps.GeoM.Reset()
	game.drawOps.GeoM.Translate(0, ScreenHeight-WallThickness)
	screen.DrawImage(game.wall[1].pict, &game.drawOps)
	game.drawOps.GeoM.Reset()
	game.drawOps.GeoM.Translate(ScreenWidth-WallThickness, 0)
	screen.DrawImage(game.wall[3].pict, &game.drawOps)

	// maze walls
	for i := 0; i < game.level[level].maxWall; i++ {
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.level[level].mazeWall[i].xLoc), float64(game.level[level].mazeWall[i].yLoc))
		screen.DrawImage(game.level[level].mazeWall[i].pict, &game.drawOps)
	}

}

// ----------------------------------------------------- End of Draw and its Functions --------------------------------

func (game Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	// database initialization
	defer gameDB.Close()
	Create_tables(gameDB)
	// database initialization end

	ebiten.SetWindowTitle(GameTitle)
	ebiten.SetWindowSize(ScreenWidth, TotalScreenHeight)

	gameObject := Game{}
	loadImage(&gameObject)
	rand.Seed(time.Now().UnixNano())

	gameObject.playerSprite.xLoc = xStart
	gameObject.playerSprite.yLoc = yStart
	gameObject.playerSprite.activeShot = false
	gameObject.playerSprite.Weapon.enemyShot = false
	gameObject.playerSprite.lives = 3
	gameObject.counter = 0

	enemyWidth, enemyHeight := gameObject.khaiSprite[0].pict.Size()

	for i := 0; i < numEnemies; i++ {
		min := 50
		maxHeight := ScreenHeight - enemyHeight - WallThickness
		maxWidth := ScreenWidth - enemyWidth - WallThickness
		xKhai = rand.Intn(maxWidth-min) + min
		yKhai = rand.Intn(maxHeight-min) + min
		gameObject.khaiSprite[i].xLoc = xKhai
		gameObject.khaiSprite[i].yLoc = yKhai
		xSophia = rand.Intn(maxWidth-min) + min
		ySophia = rand.Intn(maxHeight-min) + min
		gameObject.sophiaSprite[i].xLoc = xSophia
		gameObject.sophiaSprite[i].yLoc = ySophia

		gameObject.khaiSprite[i].hitWall = false
		gameObject.khaiSprite[i].alive = true
		gameObject.khaiSprite[i].lives = 2
		gameObject.khaiSprite[i].activeShot = false
		gameObject.khaiSprite[i].alive = true
		gameObject.sophiaSprite[i].hitWall = false
		gameObject.sophiaSprite[i].alive = true
		gameObject.sophiaSprite[i].lives = 1
		gameObject.sophiaSprite[i].activeShot = false
		gameObject.sophiaSprite[i].alive = true
	}

	if err := ebiten.RunGame(&gameObject); err != nil {
		log.Fatal("Game not running", err)
	}

} // end of main

func setImage(path string) *ebiten.Image {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("image path does not exist")
	}
	image, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		log.Fatal("setImage Error!", err)
	}
	return image
}

func loadImage(game *Game) {
	game.playerSprite.pict = setImage("images\\jackcharacter.png")
	game.playerSprite.Weapon.pict = setImage("images\\frisbee.png")
	for i := 0; i < numEnemies; i++ {
		game.khaiSprite[i].pict = setImage("images\\dragonkhai.png")
		game.khaiSprite[i].Weapon.pict = setImage("images\\watergun.png")
		game.sophiaSprite[i].pict = setImage("images\\ninjaphia.png")
		game.sophiaSprite[i].Weapon.pict = setImage("images\\watergun.png")
	}

	setWindowWall(game)
	setMaze(game)
}

func makeWallPict(width int, height int) *ebiten.Image {
	wall := ebiten.NewImage(width, height)
	wall.Fill(colornames.Cyan)
	return wall
}

func setWindowWall(game *Game) {
	game.wall[0].pict = makeWallPict(ScreenWidth, WallThickness)  // top
	game.wall[1].pict = makeWallPict(ScreenWidth, WallThickness)  // bottom
	game.wall[2].pict = makeWallPict(WallThickness, ScreenHeight) // left
	game.wall[3].pict = makeWallPict(WallThickness, ScreenHeight) // right
}

func setWallLocation(x int, y int) (xD int, yD int) {
	return x, y
}

func getWallLocation(wall Wall) (x int, y int) {
	return wall.xLoc, wall.yLoc
}

func setMaze(game *Game) {

	// level 1
	game.level[1].level = 1
	game.level[1].maxWall = 5
	game.level[1].mazeWall[0].pict = makeWallPict(WallThickness, 500)
	game.level[1].mazeWall[1].pict = makeWallPict(610, WallThickness)
	game.level[1].mazeWall[2].pict = makeWallPict(WallThickness, 360)
	game.level[1].mazeWall[3].pict = makeWallPict(WallThickness, 360)
	game.level[1].mazeWall[4].pict = makeWallPict(WallThickness, 150)
	game.level[1].mazeWall[0].xLoc, game.level[1].mazeWall[0].yLoc = setWallLocation(200, 0)
	game.level[1].mazeWall[1].xLoc, game.level[1].mazeWall[1].yLoc = setWallLocation(200, 200)
	game.level[1].mazeWall[2].xLoc, game.level[1].mazeWall[2].yLoc = setWallLocation(400, 390)
	game.level[1].mazeWall[3].xLoc, game.level[1].mazeWall[3].yLoc = setWallLocation(600, 200)
	game.level[1].mazeWall[4].xLoc, game.level[1].mazeWall[4].yLoc = setWallLocation(810, 390)

	// level 2
	game.level[2].level = 2
	game.level[2].maxWall = 6
	game.level[2].mazeWall[0].pict = makeWallPict(800, WallThickness)
	game.level[2].mazeWall[0].xLoc, game.level[2].mazeWall[0].yLoc = setWallLocation(0, 200)
	game.level[2].mazeWall[1].pict = makeWallPict(WallThickness, 180)
	game.level[2].mazeWall[1].xLoc, game.level[2].mazeWall[1].yLoc = setWallLocation(790, 200)
	game.level[2].mazeWall[2].pict = makeWallPict(WallThickness, 180)
	game.level[2].mazeWall[2].xLoc, game.level[2].mazeWall[2].yLoc = setWallLocation(590, 370)
	game.level[2].mazeWall[3].pict = makeWallPict(WallThickness, 180)
	game.level[2].mazeWall[3].xLoc, game.level[2].mazeWall[3].yLoc = setWallLocation(390, 200)
	game.level[2].mazeWall[4].pict = makeWallPict(WallThickness, 180)
	game.level[2].mazeWall[4].xLoc, game.level[2].mazeWall[4].yLoc = setWallLocation(190, 370)
	game.level[2].mazeWall[5].pict = makeWallPict(610, WallThickness)
	game.level[2].mazeWall[5].xLoc, game.level[2].mazeWall[5].yLoc = setWallLocation(190, 540)

	// level 3
	game.level[3].level = 3
	game.level[3].maxWall = 6
	game.level[3].mazeWall[0].pict = makeWallPict(600, WallThickness)
	game.level[3].mazeWall[0].xLoc, game.level[3].mazeWall[0].yLoc = setWallLocation(200, 200)
	game.level[3].mazeWall[1].pict = makeWallPict(WallThickness, 360)
	game.level[3].mazeWall[1].xLoc, game.level[3].mazeWall[1].yLoc = setWallLocation(200, 200)
	game.level[3].mazeWall[2].pict = makeWallPict(WallThickness, 350)
	game.level[3].mazeWall[2].xLoc, game.level[3].mazeWall[2].yLoc = setWallLocation(790, 200)
	game.level[3].mazeWall[3].pict = makeWallPict(400, WallThickness)
	game.level[3].mazeWall[3].xLoc, game.level[3].mazeWall[3].yLoc = setWallLocation(200, 550)
	game.level[3].mazeWall[4].pict = makeWallPict(WallThickness, 210)
	game.level[3].mazeWall[4].xLoc, game.level[3].mazeWall[4].yLoc = setWallLocation(400, 350)
	game.level[3].mazeWall[5].pict = makeWallPict(210, WallThickness)
	game.level[3].mazeWall[5].xLoc, game.level[3].mazeWall[5].yLoc = setWallLocation(590, 350)

	// game starting window
	game.level[0].level = 0
	game.level[0].maxWall = 1
	game.level[0].mazeWall[0].pict = makeWallPict(ScreenWidth, WallThickness)
	game.level[0].mazeWall[0].xLoc, game.level[0].mazeWall[0].yLoc = setWallLocation(0, 200)

	// end game
	game.level[4].level = 4

	game.currentLevel = 0
}
