package clusterdata

import org.apache.spark.sql.{SparkSession, DataFrame}
import org.apache.spark.sql.functions._
import java.nio.file.{Files, Paths, FileVisitOption}

object Task10ProcData2019 {
    def execute(onCloud: Boolean) {
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
            println("ERROR")
        } else {
            taskUsageDF = sk.read.option("compression", "gzip").json("./data/instance_usage-000000000000.json.gz")
        }
        
        taskUsageDF.printSchema()

        // 根据 alloc_collection_id 和 alloc_instance_index 进行分组，并构造每一组的 JSON 结构
        val resultDF = taskUsageDF.groupBy("alloc_collection_id", "alloc_instance_index").agg(
            collect_list(struct("start_time", "end_time", "average_usage")).as("usage_list")
        )

        resultDF.show()
        // Shut down SparkSession
        sk.stop()
    }
}