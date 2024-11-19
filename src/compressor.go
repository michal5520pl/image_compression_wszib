package main

import (
    "image"
    "image/jpeg"
    "image/png"
    "image/bmp"
    "os"
)


//funkcja rzyjmuje ścieżkę do pliku wejściowego i ścieżke do pliku zwracanego
//przyjmuje formatach PNG i BMP - zapisuje w formacie JPEG
func CompressImage(inputPath string, outputPath string, quality int) error {

    // przyjmuje plik wejściowy
    inputFile, err := os.Open(inputPath)
    if err != nil {
        return err
    }
    defer inputFile.Close()

    //dekoduje obraz
    var img image.Image
    switch ext := getFileExtension(inputPath); ext {
    case ".png":
        img, err = png.Decode(inputFile)
    case ".bmp":
        img, err = bmp.Decode(inputFile)
    default:
        return fmt.Errorf("unsupported file format: %s", ext)
    }

    if err != nil {
        return err
    }

    //inicjuje plik wyjściowy
    outputFile, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer outputFile.Close()

    //ustawienia kompresji dla formatu JPEG
    jpegOptions := jpeg.Options{Quality: quality}

    //zapisuje obraz w JPEG
    err = jpeg.Encode(outputFile, img, &jpegOptions)
    if err != nil {
        return err
    }

    return nil
}

//funkcja pobiera rozszenie pliku
func getFileExtension(filename string) string {
    if len(filename) < 4 {
        return ""
    }
    return filename[len(filename)-4:]
}