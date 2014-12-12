var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngRRKDailyProdController', ['$scope', '$http', function($scope, $http) {
  var api = "/api/rrkDailyProdEmailSendApi";
  var postData = null;

  $scope.removeEntry = function(index) {
    $scope.items.splice(index, 1);
  }

  $scope.addSingleEntry = function() {
    l = $scope.items.length;
    $scope.items.splice(l, 0, angular.copy($scope.entry));
  }

  $scope.submitTodaysLog = function() {
    $scope.statusNote = "Sending...";
    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
    }).error(function(data, status, headers, config){
      $scope.statusNote = "There was an error. Thats all there is to it. Please try again after some time";
    });
  }

  $scope.dateValue =  new Date();
  $scope.entry = {modelName:"a", quantity:5, unit:"pc", remarks:"none"};
  $scope.items = [];
  $scope.statusNote = "";

}]);

