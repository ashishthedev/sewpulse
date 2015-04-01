var bomServices = angular.module('bomServices', ['ngResource']);
bomServices.config(['$resourceProvider', function($resourceProvider) {
  // Don't strip trailing slashes from calculated URLs
  $resourceProvider.defaults.stripTrailingSlashes = false;
}]);

bomServices.factory('BOM', ['$resource', function($resource){
    return $resource('/api/bom', {}, {
      'reset': {method: 'POST', url: '/api/bom/reset'},
      'resetToSampleState': {method: 'POST', url: '/api/bom/resetToSampleBOM'}
    });
  }]);

bomServices.factory('Model', ['$resource', function($resource){
    return $resource('/api/bom/model');
  }]);

bomServices.factory('Article', ['$resource', function($resource){
    return $resource('/api/bom/article/' );
  }]);

