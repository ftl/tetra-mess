package data

import "image/color"

var (
	NoGANColor     = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	GANMinus2Color = color.RGBA{R: 139, G: 0, B: 0, A: 255}
	GANMinus1Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}
	GAN0Color      = color.RGBA{R: 255, G: 140, B: 0, A: 255}
	GAN1Color      = color.RGBA{R: 255, G: 215, B: 0, A: 255}
	GAN2Color      = color.RGBA{R: 154, G: 205, B: 50, A: 255}
	GAN3Color      = color.RGBA{R: 34, G: 139, B: 34, A: 255}
	GAN4Color      = color.RGBA{R: 0, G: 100, B: 0, A: 255}
)

func GANToColor(gan int) color.Color {
	colors := []color.Color{
		GANMinus2Color,
		GANMinus1Color,
		GAN0Color,
		GAN1Color,
		GAN2Color,
		GAN3Color,
		GAN4Color,
	}
	if gan <= NoGAN || gan > 4 {
		return NoGANColor
	}
	return colors[gan+2]
}
