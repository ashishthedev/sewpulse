var bomViewControllers = angular.module('bomViewControllers', []);

bomViewControllers.controller('ArticleDetailCtrl', ['$http', '$scope', '$routeParams', 'BOM', function($http, $scope, $routeParams, BOM) {
  var articleName = $routeParams.id;
  function FetchBom() {
    $scope.statusNote = "Fetching BOM ...";
    $scope.bom = BOM.get(
      function(data){
        $scope.statusNote = "";
        $scope.article = $scope.bom.AML.Articles[articleName];
        $scope.bomFetched = true;
      },
      function(error){
        $scope.statusNote = error.status + ": " + error.data;
      });
  }

  $scope.SaveArticleOnServerAlongWithBOM = function() {
    $scope.statusNote = "Saving changes on server";
    $scope.bom.$save(function(data){
      $scope.statusNote = "";
    }, function(error){
      $scope.statusNote = error.status + ": " + error.data;
    });
  }

  $scope.deleteArticle = function() {
    $scope.statusNote = "Deleting article on server...";
    var api = "/api/bom/article/delete";
    var api = "/api/bom/article/" + $scope.article.Name;
    var postData = $scope.article;
    $http.delete(api).success(function(data, status, headers, config) {
      $scope.statusNote = "Article " + $scope.article.Name + " deleted...";
      $scope.articleDeleted = true;
    }).error(function(data, status, headers, config) {
      $scope.statusNote = status + ": " + data;
    });
  }

  $scope.allDone = function()
  {
    return $scope.bomFetched;
  }
  function InitApp(){
    $scope.bomFetched = false;

    FetchBom();
  }
  InitApp();

}]);

bomViewControllers.controller('AdminBOMViewController', ['$scope', '$rootScope', 'BOM', function($scope, $rootScope, BOM) {
  function FetchBom() {
    $scope.statusNote = "Fetching BOM ...";
    $scope.bom = BOM.get(
      function(data){
        $scope.statusNote = "";
        var keys = [];
        for(var key in $scope.bom.AML.Articles){
          keys.push(key);
        }
        $scope.sortedArticleKeys = keys.sort();
        $scope.bomFetched = true;
      },
      function(error){
        $scope.statusNote = error.status + ": " + error.data;
      });
  }


  $scope.SaveTheChangesOnServer = function() {
    $scope.statusNote = "Saving changes on server";
    $scope.bom.$save(function(data){
      $scope.statusNote = "";
      $scope.SetGridPristine();
    }, function(error){
      $scope.statusNote = error.status + ": " + error.data;
    });
  }

  $scope.RemoveThisModel = function(ModelName){
    delete $scope.bom.Models[ModelName];
    $scope.SetGridDirty();
  }

  $scope.SetGridPristine = function() {
    $rootScope.isGridDirty = false;
  }
  $scope.SetGridDirty = function() {
    $rootScope.isGridDirty = true;
  }

  $scope.allDone = function()
  {
    return $scope.bomFetched;
  }

  function InitApp(){
    $scope.bomFetched = false;

    FetchBom();
    $scope.SetGridPristine();
  }
  InitApp()
}]);
