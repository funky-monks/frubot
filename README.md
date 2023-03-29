# frubot

Provides some auxilliary commands for use in combination with [esmbot](https://github.com/esmBot/esmBot)

## Commands
All of the image based commands require an images directory with pictures for each of the people mentioned.

### &tony
Returns an Anthony Kiedis image, analog to &cat

### &flea
Returns a Flea image

### &chris
Returns a Chris Cornell image

### &fru
Returns a John Frusciante image

### &chad
Returns a Chad Smith image

### &center
Centers the replied to image around the first found nose in the picture. Use it combination with &haah or &hooh

## Build
`go build`

## Run

`frubot -i ${imagePath} -t ${botToken}`

