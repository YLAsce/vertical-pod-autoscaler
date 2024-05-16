package clusterdata

import org.apache.spark.sql.{SparkSession, DataFrame}
import org.apache.spark.sql.functions._
import java.nio.file.{Files, Paths, FileVisitOption}

object Task7CorrelationHighConsumEviction {
    def execute(onCloud: Boolean) = {
        // Initialize schemas
        val schema_task_events = ReadSchema.read("task_events")
        schema_task_events.printTreeString()
        val schema_task_usage = ReadSchema.read("task_usage")
        schema_task_usage.printTreeString()

        // Initialize SparkSession
        var skb = SparkSession.builder().appName("CHCE")
        if(!onCloud) {
            println("Not run on cloud")
            skb = skb.master("local[*]")
        }
        val sk = skb.getOrCreate()

        // Set level of log to ERROR
        sk.sparkContext.setLogLevel("ERROR")

        // Load data files, from local machine or Google Cloud
        var taskEventsDF : DataFrame = sk.emptyDataFrame
        
        if(onCloud) {
            taskEventsDF = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .option("compression", "gzip")
                                  .schema(schema_task_events)
                                  .load("gs://clusterdata-2011-2/task_events/*.csv.gz")
        } else {
          taskEventsDF = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .schema(schema_task_events)
                                  .load("./data/task_events/*.csv")
        }

        var taskUsageDF : DataFrame = sk.emptyDataFrame
        
        if(onCloud) {
            taskUsageDF = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .option("compression", "gzip")
                                  .schema(schema_task_usage)
                                  .load("gs://clusterdata-2011-2/task_usage/*.csv.gz")
        } else {
            taskUsageDF = sk.read
                                  .format("csv")
                                  .option("header", "false")
                                  .schema(schema_task_usage)
                                  .load("./data/task_usage/*.csv")
        }

        // Extract the dataframe of the evicted tasks
        val evictedTasksDF = taskEventsDF.filter(col("Event type") === 2).select("job ID", "task index", "machine ID")
        // Extract the dataframe of the machines by the descending order of CPU consumption
        val taskUsageOrderedDF = taskUsageDF.select("machine ID", "CPU rate").orderBy(col("CPU rate").desc).distinct()
        
        // Define the threshold of "High consumption"
        val highConsumRate = 0.0001
        // Number of machines
        val numMachines = taskUsageOrderedDF.count()

        // Number of high-consuming machines
        val top : Int = (numMachines * highConsumRate).toInt
        // Extract the dataframe of high-consuming machines
        val topConsumingMachines = taskUsageOrderedDF.limit(top)

        // Combine the dataframes, extract the dataframe of high-consuming and evicted tasks
        val combinedEvictHighComsumDF = topConsumingMachines.join(evictedTasksDF, "machine ID")
                                                          .select("job ID", "task index", "machine ID")
                                                          .distinct()

        // Print logs
        println("Total High comusing machines: " + top + ", ")
        println("Total machines: " + numMachines + ", ")
        println("Total Evicted Tasks: " + evictedTasksDF.count() + ", ")
        println("Evicted Task on High Consuming machines: " + combinedEvictHighComsumDF.count())

        // Shut down SparkSession
        sk.stop()
  }
}