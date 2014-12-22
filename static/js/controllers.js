var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngRRKDailyProdController', ['$scope', '$http', function($scope, $http) {

  function UpdateTotalQty() {
    var t = 0;
    for (var i=0; i < $scope.items.length; i++) {
      t += parseInt($scope.items[i].quantity);
      console.log(t);
    }
    $scope.totalQuantityProduced = t;
  }

  $scope.removeEntry = function(index) {
    $scope.items.splice(index, 1);
    UpdateTotalQty();
  }

  $scope.addSingleEntry = function() {
    var l = $scope.items.length;
    $scope.items.splice(l, 0, angular.copy($scope.entry));
    UpdateTotalQty();
  }

  $scope.DateChanged = function() {
    var today = new Date();
    if (today < $scope.dateValue) {
      $scope.dateValue = today;
    }
  }

  $scope.submitTodaysLog = function() {
    $scope.statusNote = "Submitting...";
    var api = "/api/rrkDailyProdEmailSendApi";
    var postData = {
      "dateTimeAsUTCMilliSeconds": $scope.dateValue.getTime(),
      "items": $scope.items,
    }

    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
      $scope.isLogSubmitted = true;
    }).error(function(data, status, headers, config){
      $scope.statusNote = status + ": " + data;
      $scope.isLogSubmitted = false;
    });
  }

  $scope.dateValue =  new Date();
  $scope.entry = {modelName:"Premier Plus", quantity:5, unit:"pc", remarks:"remarks"};
  $scope.items = [];
  $scope.statusNote = "";
  $scope.isLogSubmitted = false;
  $scope.totalQuantityProduced = 0;

}]);

