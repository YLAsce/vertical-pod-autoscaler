package clusterdata

import org.apache.spark.sql.SparkSession
import org.apache.spark.SparkContext
import org.apache.spark.SparkConf
import org.apache.spark.rdd.RDD
import org.apache.commons.math3.ml.clustering.Cluster
import java.nio.file.{Files, Paths, FileVisitOption}

// Task3: What is the distribution of the number of jobs/tasks per scheduling class?
object Task3DistributionOfSchedulingClass {
   def execute(onCloud: Boolean) = {
      // start spark with 1 worker thread
      var conf = new SparkConf().setAppName("ClusterData")
      if(!onCloud) {
         println("Not run on cloud")
         conf = conf.setMaster("local[1]")
      }
      val sc = new SparkContext(conf)
      sc.setLogLevel("ERROR")

      // read the input file into an RDD, from local machine or Google Cloud
      var job_events_input = sc.textFile("./data/job_events/*.csv")
      if(onCloud) {
         job_events_input = sc.textFile("gs://clusterdata-2011-2/job_events/*.csv.gz")
      }
      val job_events_rdd = job_events_input.map(x => x.split(","))

      var task_events_input = sc.textFile("./data/task_events/*.csv")
      if(onCloud) {
         task_events_input = sc.textFile("gs://clusterdata-2011-2/task_events/*.csv.gz")
      }
      val task_events_rdd = task_events_input.map(x => x.split(","))


      // Discard the rows with null value in the column of "Scheduling class"
      val total_job_events_rdd = job_events_rdd.filter(x => x(5) != "")

      val total_task_events_rdd_var = task_events_rdd.filter(x => x(7) != "")

      // Set the key as (job ID, task index) to stand for one task
      // And set the value with (Scheduling class) only
      // At last, drop the duplicate rows
      val total_task_events_rdd = total_task_events_rdd_var.map(x => ((x(2), x(3)), x(7))).distinct()

      // Count and Print logs
      println("Distribution of Jobs: ")
      total_job_events_rdd.map(x => (x(5), 1)).reduceByKey((x, y) => (x + y)).collect().foreach(
         x => {
            println(s"- Scheduling type: ${x._1}, Number of jobs: ${x._2}")
         }
      )

      println("Distribution of Tasks: ")
      total_task_events_rdd.map(x => (x._2, 1)).reduceByKey((x, y) => (x + y)).collect().foreach(
         x => {
            println(s"- Scheduling type: ${x._1}, Number of tasks: ${x._2}")
         }
      )

      sc.stop()
   }
}