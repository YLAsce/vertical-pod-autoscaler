package clusterdata

import org.apache.spark.sql.{SparkSession, DataFrame}
import org.apache.spark.sql.functions._
import java.nio.file.{Files, Paths, FileVisitOption}

object Task9CorrFlowTimeAndConsum {
    def execute(onCloud: Boolean) {
        // Initialize schemas
        val schema_task_usage = ReadSchema.read("task_usage")
        schema_task_usage.printTreeString()

        // Initialize SparkSession
        var skb = SparkSession.builder().appName("CFTAC")
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
        
        // Create the new column for the task's working time
        val taskUsageTimeDF = taskUsageDF.withColumn("Flow time", col("end time") - col("start time"))

        // Compute the average value for the usage of CPU, usage of Memory and time consumption
        val taskUsageTimeSumDF = taskUsageTimeDF.select("job ID", "task index", "CPU rate", "canonical memory usage", "Flow time")
                                                .groupBy("job ID", "task index")
                                                .agg(avg("CPU rate").alias("SUM CPU Usage"),
                                                    avg("canonical memory usage").alias("SUM Mem Usage"),
                                                    avg("Flow time").alias("Sum Time consum"))

        // Compute the correalation coefficient between time consumption and CPU consumption
        val corrTimeCPU = taskUsageTimeSumDF.stat.corr("Sum Time consum", "SUM CPU Usage")
        // Compute the correalation coefficient between time consumption and Memory consumption
        val corrTimeMem = taskUsageTimeSumDF.stat.corr("Sum Time consum", "SUM Mem Usage")

        // Print logs
        println("Correlation between Flow time and CPU usage: " + corrTimeCPU)
        println("Correlation between Flow time and Memory usage: " + corrTimeMem)

        // Shut down SparkSession
        sk.stop()
    }
}