Date.prototype.addHours = function(h) {
   this.setTime(this.getTime() + (h*60*60*1000));
   return this;
}

Date.prototype.StartDate = function () {
  return new Date(this.getFullYear(), this.getMonth(), this.getDate());
};

Date.prototype.EndDate = function () {
  return new Date(this.getFullYear(), this.getMonth(), this.getDate(), 23, 59);
};

$('#title').html(function(){
  return window.location.host+window.location.pathname;
});

$('#number_phone').autocomplete({
  source: function( request, response ) {
      $.ajax( {
          method: 'POST',
          url: ("autocomplete_number/" + $('#number_phone').val()),
          dataType: 'json',
          data: {
            term: request.term
          },
          success: function( data ) {
            response( data );

          }
      });
    },
  minLength: 2
});

$(function() {
    $('.error').hide();
    $('input').off('keyup keypress');
  });

$('.Find').click(function() {

  phoneNum = $('#number_phone').val();
  before = $('.dataBefore').val();
  after = $('.dataAfter').val();
  //console.log(after);
  //console.log(before);
  if (phoneNum) {

  } else {
    $('.ui-widget-content.error').text('Введите номер телефона');
    $('.error').show();
    $('.error').click(function(){
      $('.error').hide();
    });
    return;
  }
  if(before) {

  } else {
    $('.ui-widget-content.error').text('Необходима начальная дата звонка');
    $('.error').show();
    $('.error').click(function(){
      $('.error').hide();
    });
    return;
  }
  if(after) {

  } else {
    $('.ui-widget-content.error').text('Необходима конечная дата звонка');
    $('.error').show();
    $('.error').click(function(){
      $('.error').hide();
    });
    return;
  }
  tmp1 = new Date(after);
  //console.log(tmp1);
  tmp2 = new Date(before);
  if(tmp1 - tmp2 < 0) {
    $('.ui-widget-content.error').text('Проверьте даты.');
    $('.error').show();
    $('.error').click(function(){
      $('.error').hide();
    });
    return;
  }
  $('#table').val(function() {
    $.ajax( {
      beforeSend: function() {
        $('#table').empty();
      },
      method: 'POST',
      url: ("number/" + phoneNum),
      data: JSON.stringify( {
        "Before": before,
        "After": after,
        "phoneNum": phoneNum
      } ),
      //contentType: "application/json; charset=utf-8",
      //processData: false,
      dataType: "html",
      success: function( data ) {
        console.log("success");
        $('#table').html(data)
      }
    }).done(function(){$( "#accordion" ).accordion()});
  });
});

$('.dataBefore').val(new Date().StartDate().addHours(3).toJSON().slice(0,19));
$('.dataAfter').val(new Date().EndDate().addHours(3).toJSON().slice(0,19));
