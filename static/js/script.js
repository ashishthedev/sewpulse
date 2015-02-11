
function DateUTCToDDMMMYY(utc) {
  var d = new Date(utc);

  function pad2(n) {
    return n > 9 ? n : '0' + n;
  }
  var MONTH_AS_TEXT = {
    1: "Jan",
    2: "Feb",
    3: "Mar",
    4: "Apr",
    5: "May",
    6: "Jun",
    7: "Jul",
    8: "Aug",
    9: "Sep",
    10: "Oct",
    11: "Nov",
    12: "Dec",
  }
  var d = new Date(utc);
  var year = d.getUTCFullYear();
  var month = MONTH_AS_TEXT[d.getUTCMonth() + 1];  // months start at zero

  var day = d.getUTCDate();

  return pad2(day) + month + year;
}

function UpdateDateDiffAsText($scope) {
  var today = new Date();
  var diff = Math.floor(today.getTime() - $scope.dateValue.getTime());
  var day = 1000 * 60 * 60 * 24;

  var days = Math.floor(diff/day);

  var dateDiffFromTodayAsText = ""
    if (days == 0) {
      dateDiffFromTodayAsText = "Today";
    }
    else if (days ==1) {
      dateDiffFromTodayAsText = "1 day old";
    } else {
      dateDiffFromTodayAsText = days + " days old";
    }
  $scope.dateDiffFromTodayAsText = dateDiffFromTodayAsText;
}
