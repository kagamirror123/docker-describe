package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type ProjectInfo struct {
	Name            string
	ComposeFilePath string
	containers      []types.Container
}

func NewRootCmd(name string, dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Short: "psx",
		Use:   name,
		Run: func(cmd *cobra.Command, _ []string) {

			ctx := context.Background()
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				panic(err)
			}

			// コンテナを取得
			containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
				All: true,
			})
			if err != nil {
				panic(err)
			}

			prjInfos := []ProjectInfo{}

			// Composeのプロジェクトごとにグループ化する
			for _, container := range containers {
				prjName, ok := container.Labels["com.docker.compose.project"]
				if !ok {
					prjName = "Not Compose Containers"
				}
				composeFilePath, ok := container.Labels["com.docker.compose.project.config_files"]
				if !ok {
					composeFilePath = ""
				}

				var (
					point  int
					exists bool
				)
				// 存在チェック
				for i, v := range prjInfos {
					if v.Name == prjName {
						exists = true
						point = i
					}
				}

				if exists {
					prjInfos[point].containers = append(prjInfos[point].containers, container)
				} else {
					pInfo := ProjectInfo{
						Name:            prjName,
						ComposeFilePath: composeFilePath,
						containers:      []types.Container{container},
					}
					prjInfos = append(prjInfos, pInfo)
				}
			}

			// 出力
			for _, v := range prjInfos {
				// 文字色を赤、背景色を青
				c := color.New(color.FgBlack, color.BgYellow)
				c.Add(color.Underline)
				fmt.Printf("⭐ ProjectName : ")
				c.Println(v.Name)
				fmt.Printf("📃 ComposeFile : ")
				c.Println(v.ComposeFilePath)
				fmt.Println()

				headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
				columnFmt := color.New(color.FgYellow).SprintfFunc()

				tbl := table.New("ID", "IMAGE", "CREATED", "COMMAND", "STATUS", "PORTS", "NAMES")
				tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

				for _, container := range v.containers {
					cid := container.ID[:15]
					var image string
					if len(container.Image) < 20 {
						image = container.Image
					} else {
						image = container.Image[:20]
					}
					dtFromUnix := time.Unix(container.Created, 0)
					createdAt := units.HumanDuration(time.Now().UTC().Sub(dtFromUnix)) + " ago"
					command := container.Command[:15]
					portPublishers := []api.PortPublisher{}
					name := strings.TrimLeft(container.Names[0], "/")
					for _, p := range container.Ports {
						pp := api.PortPublisher{
							URL:           p.IP,
							TargetPort:    int(p.PrivatePort),
							PublishedPort: int(p.PublicPort),
							Protocol:      p.Type,
						}
						portPublishers = append(portPublishers, pp)
					}
					ports := DisplayablePorts(portPublishers)
					tbl.AddRow(cid, image, createdAt, command, container.Status, ports, name)
				}

				tbl.Print()

				fmt.Println("\n\n")
			}
		},
	}

	return cmd
}

type portRange struct {
	pStart   int
	pEnd     int
	tStart   int
	tEnd     int
	IP       string
	protocol string
}

func (pr portRange) String() string {
	var (
		pub string
		tgt string
	)

	if pr.pEnd > pr.pStart {
		pub = fmt.Sprintf("%s:%d-%d->", pr.IP, pr.pStart, pr.pEnd)
	} else if pr.pStart > 0 {
		pub = fmt.Sprintf("%s:%d->", pr.IP, pr.pStart)
	}
	if pr.tEnd > pr.tStart {
		tgt = fmt.Sprintf("%d-%d", pr.tStart, pr.tEnd)
	} else {
		tgt = fmt.Sprintf("%d", pr.tStart)
	}
	return fmt.Sprintf("%s%s/%s", pub, tgt, pr.protocol)
}

// DisplayablePorts is copy pasted from https://github.com/docker/cli/pull/581/files
func DisplayablePorts(p api.PortPublishers) string {
	if p == nil {
		return ""
	}

	sort.Sort(p)

	pr := portRange{}
	ports := []string{}
	for _, p := range p {
		prIsRange := pr.tEnd != pr.tStart
		tOverlaps := p.TargetPort <= pr.tEnd

		// Start a new port-range if:
		// - the protocol is different from the current port-range
		// - published or target port are not consecutive to the current port-range
		// - the current port-range is a _range_, and the target port overlaps with the current range's target-ports
		if p.Protocol != pr.protocol || p.URL != pr.IP || p.PublishedPort-pr.pEnd > 1 || p.TargetPort-pr.tEnd > 1 || prIsRange && tOverlaps {
			// start a new port-range, and print the previous port-range (if any)
			if pr.pStart > 0 {
				ports = append(ports, pr.String())
			}
			pr = portRange{
				pStart:   p.PublishedPort,
				pEnd:     p.PublishedPort,
				tStart:   p.TargetPort,
				tEnd:     p.TargetPort,
				protocol: p.Protocol,
				IP:       p.URL,
			}
			continue
		}
		pr.pEnd = p.PublishedPort
		pr.tEnd = p.TargetPort
	}
	if pr.tStart > 0 {
		ports = append(ports, pr.String())
	}
	return strings.Join(ports, ", ")
}
