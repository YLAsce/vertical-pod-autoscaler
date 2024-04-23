package clusterdata

import org.apache.spark.sql.{SparkSession, DataFrame}
import org.apache.spark.sql.functions._
import java.nio.file.{Files, Paths, FileVisitOption}

// Task5: In general, do tasks from the same job run on the same machine?
object Task5SameJobMachine {
    def execute(onCloud: Boolean) = {
        // Initialize schemas
        val schema_task_events = ReadSchema.read("task_events")

        // Initialize spark session
        var skb = SparkSession.builder().appName("SJM")
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

        // Count the number of machines each job used
        val df_job_distinct_machine_num = df_task_events_var.groupBy("job ID")
            .agg(
                countDistinct("machine ID").as("num machine used")
            )

        //Cache the dataframe, for multiple future actions
        df_job_distinct_machine_num.cache()

        //Count the amout of jobs who use certain number of machines, to see distribution
        df_job_distinct_machine_num.groupBy("num machine used")
            .count().sort("num machine used").show()

        // Result
        println("Number of Jobs in which tasks use 1 machine:")
        println(df_job_distinct_machine_num.filter(col("num machine used") === 1).count())
        println("Number of Jobs:")
        println(df_job_distinct_machine_num.count())
        
        sk.stop()
    }
}