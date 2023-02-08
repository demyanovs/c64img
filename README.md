# c64img

c64img is a generator of a BASIC program for Commodore 64 that draws a simple image on the screen using colored characters from a provided static image.

## Details

### How commodore 64 screen works

The Commodore 64 screen consists of 25 rows of 40 characters each.

The first character of the screen (top left) starts at address 1024, and each next character is represented by 1024 + 1 address.

### How c64img generator works

c64img reads provided image, converts colors to the Commodore 64 predefined palette, and for each pixel adds corresponding color code to the array.

The array of color codes is then written to the file as a BASIC program using the DATA and POKE commands to set the value to a memory address.

## Requirements

The source image must be in one of the formats: jpeg, png or gif and have a resolution of 40x25 pixels.

The image can contain any number of colors, it will be automatically converted to the Commodore 64 predefined palette.

## Installation

c64img generator is written in the go programming language and can be built with the go command:

```bash
go build .\c64img.go
```

## Usage

```bash
./c64img -i img.png -f img.basic -o out.png -dither
```

```
-i - path to the input image (required)
-f - path to the file where the generated BASIC program will be stored
-o - path to the output image comverted to Commodore 64 palette
-dither - use the Floyd–Steinberg dithering algorithm to convert the image (by default "false")
```

You can always run for help:
```bash
./c64img -help
```

## Screenshots
<p align="center">
    <img src="https://demyanov.dev/images/go/c64img/img_01_400.png" style=" width:400px;"  alt="yoda">
    <img src="https://demyanov.dev/images/go/c64img/img_02_400.png" style=" width:400px;"  alt="c64 logo">
</p>

## Credits
This generator was inspired by [64bites](https://64bites.com/blog/2015/05/31/create-a-1k-image-for-c64-with-ruby/) video series.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)