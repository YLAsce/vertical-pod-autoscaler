package clusterdata

object Main {
  def main(args: Array[String]): Unit = {
    if (args.length < 2) {
      println("[ERROR] Usage: spark-submit target/scala-2.12/clusterdata_2.12-1.0.jar [taskID] [cloud/local]")
      System.exit(1)
    }
    // Different settings when executing on local machine or on Google cloud
    val onCloud = if (args(1) == "cloud") true else false
    
    val taskName = args(0)
    // Execute task according to the input task id
    taskName match {
      case "task1" => Task1DistributionOfCPU.execute(onCloud)
      case "task2" => Task2PercentageOfCompPowerLost.execute(onCloud)
      case "task3" => Task3DistributionOfSchedulingClass.execute(onCloud)
      case "task4" => Task4SchedulingClassEviction.execute(onCloud)
      case "task5" => Task5SameJobMachine.execute(onCloud)
      case "task6" => Task6ResourceRequestConsume.execute(onCloud)
      case "task7" => Task7CorrelationHighConsumEviction.execute(onCloud)
      case "task8" => Task8TaskConsumCPUAndMem.execute(onCloud)
      case "task9" => Task9CorrFlowTimeAndConsum.execute(onCloud)
      case "task10" => Task10ProcData2019.execute(onCloud)
      // Task not exist
      case _ =>
        println(s"[ERROR] Unknown task ID: $taskName")
        System.exit(1)
    }
  }
}