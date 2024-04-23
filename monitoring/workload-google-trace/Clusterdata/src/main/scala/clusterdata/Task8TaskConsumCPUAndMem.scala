package clusterdata

import org.apache.spark.sql.{SparkSession, DataFrame}
import org.apache.spark.sql.functions._
import java.nio.file.{Files, Paths, FileVisitOption}

object Task8TaskConsumCPUAndMem {
    def execute(onCloud: Boolean) {
        // Initialize schemas
        val schema_task_usage = ReadSchema.read("task_usage")
        schema_task_usage.printTreeString()

        // Initialize SparkSession
        var skb = SparkSession.builder().appName("TCCAM")
        if(!onCloud) {
            println("Not run on cloud")
            skb = skb.master("local[*]")
        }
        val sk = skb.getOrCreate()

        // Set level of log to ERROR
        sk.sparkContext.setLogLevel("ERROR")

        // Load data files, from local machine or Google Cloud
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

        // Aggregate sum the CPU usage and Memory usage by each task
        val taskCPUConsumSumDF = taskUsageDF.select("job ID", "task index", "CPU rate")
                                            .groupBy("job ID", "task index")
                                            .agg(avg("CPU rate").alias("SUM CPU Usage"))
        val taskMemConsumSumDF = taskUsageDF.select("job ID", "task index", "canonical memory usage")
                                            .groupBy("job ID", "task index")
                                            .agg(avg("canonical memory usage").alias("SUM Mem Usage"))

        // Set the range
        val topCPU : Int = 1000
        val topMem : Int = 1000

        // Extract the DataFrame of tasks with Top N consumption
        val taskHighCPUConsumDF = taskCPUConsumSumDF.sort(col("SUM CPU Usage").desc).limit(topCPU)
        val taskHighMemConsumDF = taskMemConsumSumDF.sort(col("SUM Mem Usage").desc).limit(topMem)
        
        // Connect and find the tasks that have a high CPU/Memory consumption at the same time
        val combinedDF = taskHighCPUConsumDF.join(taskHighMemConsumDF, Seq("job ID", "task index"))

        // Print logs
        println("Number of tasks that simultaneously have TOP " 
                + topCPU + " CPU usage and TOP " 
                + topMem + " Memory usage: " 
                + combinedDF.count())
        
        // Shut down SparkSession
        sk.stop()
    }
}