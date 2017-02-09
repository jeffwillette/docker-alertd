package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
        for {
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}

		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			panic(err)
		}

		for _, container := range containers {
		        s, err := cli.ContainerStats(context.Background(), container.ID, false)
			if err != nil { panic(err) }

			var stats []byte
			buf := make([]byte, 1024)
			for {
				n, err := s.Body.Read(buf)
				stats = append(stats, buf[:n]...)
				if err != nil {
					if err != io.EOF {
						fmt.Println("read error:", err)
					}
					break
				}
			}

			var data interface{} 

			e := json.Unmarshal(stats, &data)
			if e != nil { fmt.Println(e) }

//			fmt.Println(data)
		}
	}
}
