package clusterdata

import org.apache.spark.sql.SparkSession
import org.apache.spark.SparkContext
import org.apache.spark.SparkConf
import org.apache.spark.rdd.RDD
import org.apache.commons.math3.ml.clustering.Cluster

// Task1: What is the distribution of the machines according to their CPU capacity?
object Task1DistributionOfCPU {
   def execute(onCloud: Boolean) = {
      // start spark with 1 worker thread
      var conf = new SparkConf().setAppName("ClusterData")
      if(!onCloud) {
         println("Not run on cloud")
         conf = conf.setMaster("local[1]")
      }
      val sc = new SparkContext(conf)
      sc.setLogLevel("ERROR")

      // read the input file into an RDD[String], from local machine or Google Cloud
      var machine_events_input = sc.textFile("./data/machine_events/part-00000-of-00001.csv.gz")
      if(onCloud) {
         machine_events_input = sc.textFile("gs://clusterdata-2011-2/machine_events/*.csv.gz")
      }

      val machine_events_rdd = machine_events_input.map(x => x.split(","))
      val rdd_CPU_Capacity = machine_events_rdd.map(x => (  
         try {
            // Extract the column of CPUs (5) and machine ID (2)
            // Notice that the index begins with 0 !
            val CPU_Capacity = x(4)
            val machine_ID = x(1)
            
            // Return the key-value pair 
            (CPU_Capacity, machine_ID)
         } catch {
            case e: ArrayIndexOutOfBoundsException =>
               // The row which exists "NA"
               println(s"ArrayIndexOutOfBoundsException on line: ${x.mkString(",")}")

               // Return the default pair
               ("NA", "NA")
         })
      ).groupByKey() // group the key-value pairs by the column of "CPUs"

      // The length of the vector of value 
      // is the number of machines
      // corresponding to the key (CPU capacity)
      val rdd_count = rdd_CPU_Capacity.map(x => (x._1, x._2.size)).collect()

      // Print logs
      println("Distribution of the machines according to their CPU capacity: ")
      rdd_count.foreach(x => {
         val key = x._1
         val values = x._2
         println(s"CPU Capacity: $key, Number of the machines: $values")
      })

      sc.stop()
   }
}