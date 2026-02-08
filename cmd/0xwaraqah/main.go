package main

import (
	"fmt"
	"log"
	"os"

	"github.com/0xdevar/waraqah/logic"
	"github.com/0xdevar/waraqah/repos"
	// "github.com/0xdevar/waraqah/ui/bootstrap"
)

func main() {
	cacheDir, err := os.UserCacheDir()

	if err != nil {
		log.Panicln(err)
	}

	repo, err := repos.NewGitRepo("0xdevar", "0xwaraqat", "main", fmt.Sprintf("%s/0xwaraqah", cacheDir))

	if err != nil {
		log.Panicln(err)
	}

	wallpapers, err := logic.RetrieveWallpapers(repo, 10)

	if err != nil {
		log.Panicln(err)
		os.Exit(-1)
	}

	collections := wallpapers.GetWallpapers(0)

	for _, wallpaper := range collections {
		fmt.Println(wallpaper.Name)
		fmt.Println("Images:")

		if err := repo.DownloadWallpaper(wallpaper); err != nil {
			log.Printf("cannot download [%s]\n", wallpaper.Name)
			continue
		}

		for _, image := range wallpaper.Images {
			fmt.Println(image.Path)
		}

		fmt.Println(wallpaper.Thumnail)


	}


	// bootstrap.Run()
}
