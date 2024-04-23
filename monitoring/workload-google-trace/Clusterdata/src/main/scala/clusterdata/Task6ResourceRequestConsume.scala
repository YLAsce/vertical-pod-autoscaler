package clusterdata

import org.apache.spark.sql.{SparkSession, DataFrame}
import org.apache.spark.sql.functions._
import java.nio.file.{Files, Paths, FileVisitOption}

// Task7: Can we observe correlations between peaks of high resource consumption on some machines and task eviction events?
object Task6ResourceRequestConsume {
    def execute(onCloud: Boolean) = {
        // Initialize schemas
        val schema_task_events = ReadSchema.read("task_events")
        schema_task_events.printTreeString()
        val schema_task_usage = ReadSchema.read("task_usage")
        schema_task_usage.printTreeString()

        // Initialize SparkSession
        var skb = SparkSession.builder().appName("RRC")
        if(!onCloud) {
            println("Not run on cloud")
            skb = skb.master("local[*]")
        }
        val sk = skb.getOrCreate()

        // Set level of log to ERROR
        sk.sparkContext.setLogLevel("ERROR")

        // Load data files, from local machine or Google Cloud
        var df_task_events : DataFrame = sk.emptyDataFrame
        
        if(onCloud) {
            df_task_events = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .option("compression", "gzip")
                                  .schema(schema_task_events)
                                  .load("gs://clusterdata-2011-2/task_events/*.csv.gz")
        } else {
            df_task_events = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .schema(schema_task_events)
                                  .load("./data/task_events/*.csv")
        }

        var df_task_usage : DataFrame = sk.emptyDataFrame
        if(onCloud) {
            df_task_usage = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .option("compression", "gzip")
                                  .schema(schema_task_usage)
                                  .load("gs://clusterdata-2011-2/task_usage/*.csv.gz")
        } else {
            df_task_usage = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .schema(schema_task_usage)
                                  .load("./data/task_usage/*.csv")
        }

        //Get CPU memory and disk request of all tasks
        val df_task_resource_request = df_task_events.groupBy("job ID", "task index")
            .agg(
                avg("CPU request").as("CPU request"),
                avg("memory request").as("memory request"),
                avg("disk space request").as("disk request")
            )

        //Get CPU memory and disk usage of all tasks
        val df_task_resource_usage = df_task_usage.groupBy("job ID", "task index")
            .agg(
                avg("CPU rate").as("CPU usage"),
                avg("assigned memory usage").as("memory usage"),
                avg("local disk space usage").as("disk usage")
            )

        //Join into the same table
        val df_joined_request_usage = df_task_resource_request.join(df_task_resource_usage, Seq("job ID", "task index"))
        //df_joined_request_usage.show()

        //Cache the dataframe, for multiple future actions
        df_joined_request_usage.cache()

        // Results
        println("Correlation between CPU request and usage:")
        println(df_joined_request_usage.stat.corr("CPU request", "CPU usage"))

        println("Correlation between memory request and usage:")
        println(df_joined_request_usage.stat.corr("memory request", "memory usage"))

        println("Correlation between disk request and usage:")
        println(df_joined_request_usage.stat.corr("disk request", "disk usage"))
    }
}