
function JSDateToUnixTime(jsDate){
  return Math.floor(jsDate.getTime()/1000);
}

function DDMMMYYFromUnixTime(unixTime){
  return DateAsUnixTimeToDDMMMYY(unixTime);
}
function DateAsUnixTimeToDDMMMYY(unixTime) {
  var d = new Date(unixTime);

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
  var d = new Date(unixTime*1000);
  var year = d.getUTCFullYear();
  var month = MONTH_AS_TEXT[d.getUTCMonth() + 1];  // months start at zero

  var day = d.getUTCDate();

  return pad2(day) + month + year;
}

function GetDateDiffAsText(JSDate){
  var today = new Date();
  var diff = Math.floor(today.getTime()/1000 - JSDate.getTime()/1000);
  var day = 60 * 60 * 24;

  var days = Math.floor(diff/day);

  var dateDiffFromTodayAsText = "";
  if (days == 0) {
    dateDiffFromTodayAsText = "Today";
  }
  else if (days == 1) {
    dateDiffFromTodayAsText = "1 day old";
  } else {
    dateDiffFromTodayAsText = days + " days old";
  }
  return dateDiffFromTodayAsText;

}

function UpdateDateDiffAsText($scope) {
  $scope.dateDiffFromTodayAsText = GetDateDiffAsText($scope.dateValue);
}

function Debug(){
  return true;
  return false;
}
