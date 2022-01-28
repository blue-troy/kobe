package task

import (
	"fmt"
	"log"

	"github.com/KubeOperator/kobe/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var taskDescribeCmd = &cobra.Command{
	Use: "describe",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("server.host")
		port := viper.GetInt("server.port")
		c := client.NewKobeClient(host, port)
		if len(args) < 1 {
			log.Fatal("task id missing")
		}
		taskId := args[0]
		result, err := c.GetResult(taskId)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("id: %s \n", result.Id)
		fmt.Printf("star time: %s \n", result.StartTime)
		fmt.Printf("end time: %s \n", result.EndTime)
		fmt.Printf("finished: %t \n", result.Finished)
		fmt.Printf("success: %t \n", result.Success)
		fmt.Printf("message:%s \n", result.Message)
		fmt.Printf("content: \n")
		fmt.Println(result.Content)
	},
}

func init() {

}
