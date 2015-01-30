var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngRRKDailyPolishController', ['$scope', '$http', function($scope, $http) {
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
    UpdateDateDiffAsText($scope);
  }

  $scope.submitTodaysLog = function() {
    $scope.statusNote = "Submitting...";
    var api = "/api/rrkDailyPolishEmailSendApi";
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

  $scope.createModelOnServerWithName = function() {
    $scope.statusNote = "Creating new model...";
    var api = "/api/rrkAddModelNameApi";
    var postData = {"NewModelName": $scope.newModelName};

    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
      $scope.models = data.ModelNames;
      FetchModelsFromServer();
      $scope.newModelName = "";
    }).error(function(data, status, headers, config){
      $scope.statusNote = status + ": " + data;
    });
  }

  function FetchModelsFromServer() {
    $scope.statusNote = "Fetching model set...";
    var api = "/api/rrkGetModelApi";
    var postData = {};

    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
      $scope.models = data.ModelNames;
      console.log("data.ModelNames : "+ data.ModelNames)
      if( $scope.models.length === 0){
        $scope.models[0] = "No model exists. Please create one.";
      }
      $scope.entry.modelName = $scope.models[0];
      console.log("setting the entry.Modelname to : "+ $scope.entry.modelName)
    }).error(function(data, status, headers, config){
      $scope.statusNote = status + ": " + data;
    });
  }

  $scope.dateValue =  new Date();
  UpdateDateDiffAsText($scope);
  $scope.entry = {modelName:"Premium Plus", quantity:5, assemblyLineName: "Line1", unit:"pc"};
  $scope.items = [];
  $scope.statusNote = "";
  $scope.isLogSubmitted = false;
  $scope.totalQuantityProduced = 0;
  FetchModelsFromServer();

}]);
