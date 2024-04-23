package clusterdata

import org.apache.spark.sql.SparkSession
import org.apache.spark.SparkContext
import org.apache.spark.SparkConf
import org.apache.spark.rdd.RDD
import org.apache.commons.math3.ml.clustering.Cluster

// Task2: What is the percentage of computational power lost due to maintenance (a machine went offline and reconnected later)?
object Task2PercentageOfCompPowerLost {
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
      var machine_events_input = sc.textFile("./data/machine_events/part-00000-of-00001.csv")
      if(onCloud) {
         machine_events_input = sc.textFile("gs://clusterdata-2011-2/machine_events/*.csv.gz")
      }
      val machine_events_rdd = machine_events_input.map(x => x.split(","))

      val rdd_EventType_CPUCapacity = machine_events_rdd.map(x => (  
         try {
            // Extract the column of event type (3) and CPUs (5)
            // Notice that the index begins with 0 !
            val Event_Type = x(2)
            val CPU_Capacity = x(4)
            
            // Return the key-value pair 
            (Event_Type, CPU_Capacity)
         } catch {
            case e: ArrayIndexOutOfBoundsException =>
               // The row which exists "NA"
               println(s"ArrayIndexOutOfBoundsException on line: ${x.mkString(",")}")

               // Return the default pair
               ("NA", "NA")
         })
      )

      // Discard the key-value pairs with null value
      val rdd_EventType_CPUCapacity_filtered = rdd_EventType_CPUCapacity.filter(x => x._1 != "NA" && x._2 != "NA")

      // Since the "NA" converts the type of elements in the column into String, convert them back to FLOAT
      val rdd_EventType_CPUCapacity_Convert = rdd_EventType_CPUCapacity_filtered.mapValues(value => value.toFloat)

      // Compute the computational power by the machine's event type
      val rdd_EventType_CPUCapacity_SUM = rdd_EventType_CPUCapacity_Convert.reduceByKey((x, y) => (x + y))

      // Compute the total computational power
      val total_sum = rdd_EventType_CPUCapacity_SUM.map(x => x._2).reduce(_ + _)

      // Compute the computational power of the "went offline..." machines
      val eventtype1_sum = rdd_EventType_CPUCapacity_SUM.filter(x => x._1 == "1").first()._2

      // Print logs
      println("total sum: " + total_sum + "event type 1 sum: " + eventtype1_sum)
      println("Percentage of Lost Computational Power: " + ((eventtype1_sum / total_sum) * 100.0) + "%")

      sc.stop()
   }
}