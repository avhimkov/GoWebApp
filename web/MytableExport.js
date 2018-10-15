        
 var $table = $('#table');
 $(function () {
     $('#toolbar').find('select').change(function () {
         $table.bootstrapTable('destroy').bootstrapTable({
             exportDataType: $(this).val()
         });
     });
 })