var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngRRKDailyProdController', ['$scope', '$http', function($scope, $http) {
  var api = "/api/rrkDailyProdEmailSendApi";
  var postData = null;

  $scope.statusNote = "Fetching...";
  $http.post(api, postData).success(function(data, status, headers, config) {
    $scope.statusNote = "";
  }).error(function(data, status, headers, config){
    $scope.statusNote = "There was an error. Thats all there is to it. Please try again after some time";
  });


  $scope.dateValue =  new Date();

}]);

