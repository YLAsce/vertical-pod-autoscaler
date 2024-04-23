package clusterdata

import org.apache.spark.sql.{SparkSession, DataFrame}
import org.apache.spark.sql.functions._
import java.nio.file.{Files, Paths, FileVisitOption}

// Task4: Do tasks with a low scheduling class have a higher probability of being evicted?
object Task4SchedulingClassEviction {
    def execute(onCloud: Boolean) = {
        // Initialize schemas
        val schema_task_events = ReadSchema.read("task_events")
        schema_task_events.printTreeString()

        // Initialize spark session
        var skb = SparkSession.builder().appName("SCE")
        if(!onCloud) {
            println("Not run on cloud")
            skb = skb.master("local[*]")
        }
        val sk = skb.getOrCreate()

        // Set level of log to ERROR
        sk.sparkContext.setLogLevel("ERROR")

        // Load data files, from local machine or Google Cloud
        var df_task_events_var : DataFrame = sk.emptyDataFrame

        if(onCloud) {
            df_task_events_var = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .option("compression", "gzip")
                                  .schema(schema_task_events)
                                  .load("gs://clusterdata-2011-2/task_events/*.csv.gz")
        } else {
            df_task_events_var = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .schema(schema_task_events)
                                  .load("./data/task_events/*.csv")
        }

        // Get scheduling class and if it's ever evicted for each task
        val df_class_eviction = df_task_events_var.groupBy("job ID", "task index")
            .agg(
                max(when(col("event type").equalTo(2), 1).otherwise(0)).as("evicted"),
                first("scheduling class").as("scheduling class"),
            )

        // Calculate the percent of tasks being evicted for each scheduling class
        val df_percent_eviction_by_class = df_class_eviction.groupBy("scheduling class")
            .agg(
                (sum(col("evicted")) / count("*")).as("percent_evicted")
            ).sort("scheduling class")

        df_percent_eviction_by_class.show()

        sk.stop()
    }
}