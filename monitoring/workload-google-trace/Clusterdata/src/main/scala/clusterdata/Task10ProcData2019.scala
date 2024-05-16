package clusterdata

import org.apache.spark.sql.{SparkSession, DataFrame}
import org.apache.spark.sql.functions._
import java.nio.file.{Files, Paths, FileVisitOption}

object Task10ProcData2019 {
    def execute(onCloud: Boolean, allocId: String) {
        // Initialize SparkSession
        var skb = SparkSession.builder().appName("PD")
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
            taskUsageDF = sk.read.option("compression", "gzip").json("gs://clusterdata_2019_a/instance_usage-000000000000.json.gz")
        } else {
            taskUsageDF = sk.read.option("compression", "gzip").json("./data/instance_usage-000000000000.json.gz")
            //taskUsageDF = sk.read.option("compression", "gzip").json("gs://clusterdata_2019_a/instance_usage-000000000000.json.gz")
        }
        
        taskUsageDF.printSchema()
        val resultDF = taskUsageDF.filter("alloc_collection_id == '%s'".format(allocId))
                            .select(
                                col("start_time").cast("Long"),
                                col("end_time").cast("Long"),
                                col("alloc_instance_index"),
                                col("collection_id"),
                                col("instance_index"),
                                col("machine_id"),
                                col("average_usage.cpus"),
                                col("average_usage.memory")
                                )
                            .orderBy("start_time")
        // 根据 alloc_collection_id 和 alloc_instance_index 进行分组，并构造每一组的 JSON 结构
        // val resultDF = taskUsageDF.groupBy("alloc_collection_id", "alloc_instance_index").agg(
        //     collect_list(struct("start_time", "end_time", "average_usage")).as("usage_list")
        // )

        resultDF.show()
        if(onCloud) {
            resultDF.write.json("gs://clusterdata_autopilot/%s_output.json".format(allocId))
        } else {
            resultDF.write.json("./data/%s_output.json".format(allocId))
            //taskUsageDF = sk.read.option("compression", "gzip").json("gs://clusterdata_2019_a/instance_usage-000000000000.json.gz")
        }
        
        // Shut down SparkSession
        sk.stop()
    }
}