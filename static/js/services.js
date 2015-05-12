var bomServices = angular.module('bomServices', ['ngResource']);

bomServices.factory('BOM', ['$resource', function($resource){
    return $resource('/api/bom', {}, {
      'reset': {method: 'POST', url: '.a/api/bom/reset'},
      'resetToSampleState': {method: 'POST', url: '/a/api/bom/resetToSampleBOM'}
    });
  }]);

bomServices.factory('Model', ['$resource', function($resource){
    return $resource('/api/bom/model/:id',{},{
      'query':{method:'GET', isArray:false}
    });
  }]);

bomServices.factory('Article', ['$resource', function($resource){
    return $resource('/api/bom/article/:id',{},{
      'query':{method:'GET', isArray:false}
    });
  }]);

bomServices.factory('RRKSaleInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/saleInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKPurchaseInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/purchaseInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKFPOutwardStkTrfrInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/fpOutwardStkTrfInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKFPInwardStkTrfrInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/fpInwardStkTrfInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKFPAdhocAdjInvoice', ['$resource', function($resource){
    return $resource('/a/api/rrk/fpAAInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKRMAdhocAdjInvoice', ['$resource', function($resource){
    return $resource('/a/api/rrk/rmAAInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKRMInwardStkTrfrInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/rmInwardStkTrfInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKRMOutwardStkTrfrInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/rmOutwardStkTrfInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKStockPosition', ['$resource', function($resource){
    return $resource('/api/rrk/stock-position-for-date/:id',{},{
    } );
  }]);

bomServices.factory('RRKStockPristineDate', ['$resource', function($resource){
    return $resource('/api/rrk/stock-pristine-date',{},{
    } );
  }]);
