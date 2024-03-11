package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

func WriteHeap() {
	log.Info().Msg("Writing  profiles")
	resp, err := http.Get("http://localhost:8080/debug/pprof/heap")
	if err != nil {
		log.Info().Msgf("Failed at http.Get error: %+v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Info().Msgf("Failed at read body error: %+v", err)
		return
	}

	now := time.Now()
	dirName := "/tmp/goprofile"
	if _, err := os.Stat(dirName); err != nil {
		err := os.MkdirAll(dirName, os.ModePerm)
		if err != nil {
			log.Err(err).Msgf("Failed to create dir: %s", dirName)
			return
		}
	}
	fileName := fmt.Sprintf("%s/heap.%d.pprof", dirName, now.UnixMilli())
	err = os.WriteFile(fileName, body, 0o644)
	if err != nil {
		log.Info().Msgf("Failed at write body error: %+v", err)
		return
	}
	log.Info().Msgf("Wrote heap file: %s", fileName)
}
