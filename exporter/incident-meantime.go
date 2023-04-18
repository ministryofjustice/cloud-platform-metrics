package exporter

import (
	"flag"
	"fmt"
)

func incidentmeantime() ([]map[string]float64, error) {
	flag.Parse()
	infraReport := make([]map[string]float64, 0)

	infraPRMap := make(map[string]float64)

	infraPRMap["incidents_mean_time_to_repair"] = 225.11
	infraPRMap["incidents_mean_time_to_resolve"] = 225.28
	infraReport = append(infraReport, infraPRMap)
	return infraReport, nil
}
func hours_and_minutes(inSeconds int) string {
	minutes := inSeconds / 60
	seconds := inSeconds % 60
	str := fmt.Sprintf("d:d", minutes, seconds)
	return str
}

func main() {
	fmt.Println("3600 seconds in minutes : ", hours_and_minutes(3600))
	fmt.Println("9999 seconds in minutes : ", hours_and_minutes(9999))
	fmt.Println("660 seconds in minutes : ", hours_and_minutes(660))
	fmt.Println("1234567890 seconds in minutes : ", hours_and_minutes(1234567890))
	// puts parse_incident_log
	//
	//	.map { |quarter, incidents| Quarter.new(quarter, incidents) }
	//	.sort { |a, b| a.title <=> b.title }
	//	.map(&:report)
}
func parse_incident_log() {
	//  data = {}
	//  current_quarter = nil
	//  current_incident = nil

	//  IO.foreach(INCIDENT_LOG) do |l|
	//    line = l.chomp

	//    case line
	//    when QUARTER_REGEX
	//      if current_quarter # i.e. this isn't the first quarter marker in the file
	//        # We just reached the next quarter, so add the current incident to the current quarter
	//        data[current_quarter].push(current_incident) unless current_incident.nil?
	//      end

	//      # initialise a new 'quarter' hash
	//      current_quarter = line
	//      data[current_quarter] = []
	//    when INCIDENT_REGEX
	//      # We reached an incident marker. Finish off this incident and start a new hash.
	//      data[current_quarter].push(current_incident) unless current_incident.nil?
	//      current_incident = {}
	//    when TIME_TO_REPAIR_REGEX
	//      # We've found the time_to_repair line for this incident
	//      m = TIME_TO_REPAIR_REGEX.match(line)
	//      current_incident[:time_to_repair] = m[1]
	//    when TIME_TO_RESOLVE_REGEX
	//      # We've found the time_to_resolve line for this incident
	//      m = TIME_TO_RESOLVE_REGEX.match(line)
	//      current_incident[:time_to_resolve] = m[1]
	//    end
	//  end

	//  # Ensure we handle the last incident in the file
	//  data[current_quarter].push(current_incident) unless current_incident == {}

	//	data
	//
	// end
}

func report() {
	//  def report
	//    <<~EOF
	//    #{title}
	//      Incidents: #{incidents.length}
	//      Mean time to repair: #{mean_time_to_repair}
	//      Mean time to resolve: #{mean_time_to_resolve}
	//    EOF
	//  end

	// private
}

func mean_time_to_repair() {
	//  def mean_time_to_repair
	//    sum = incidents.map(&:time_to_repair).sum
	//    hours_and_minutes(sum / incidents.length)
	//  end

}

func mean_time_to_resolve() {
	// def mean_time_to_resolve
	//
	//	sum = incidents.map(&:time_to_resolve).sum
	//	hours_and_minutes(sum / incidents.length)
	//
	// end
}

func initialise() {
	var (
	//              INCIDENT_LOG = "incident-log.html.md.erb"

	//              QUARTER_REGEX         = ""
	//              INCIDENT_REGEX        = ""
	//              TIME_TO_REPAIR_REGEX  = ""
	//              TIME_TO_RESOLVE_REGEX = ""
	)
	//class Quarter
	//  attr_reader :title, :incidents

	//	 def initialize(title, incidents)
	//	   @title = title
	//	   @incidents = incidents
	//	     .reject { |i| i == {} }
	//	     .map { |i| Incident.new(i) }
	//	 end
	//		return
}

func time_to_resolve() {
	//  def time_to_resolve
	//    to_seconds(@time_to_resolve || @time_to_repair) # In case one is missing
	//  end

}

func time_to_repair() {
	//  def time_to_repair
	//    to_seconds(@time_to_repair || @time_to_resolve) # In case one is missing
	//  end

}

func to_seconds() {
	//	def to_seconds(str)
	//	  hours, minutes = str.split(" ").map(&:to_i)
	//	  (hours * 3600) + (minutes * 60)
	//	end
	//
	// end
}
