var main = function() {
  $('.thumbnail').click(function(){
    $(this).toggleClass('faded');

    if ($(this).hasClass('faded')) {
      $(this).attr('active', false);
      console.log("now faded "+ $(this).attr('active'));
    } else {
      $(this).attr('active', true);
      console.log("not faded "+ $(this).attr('active'));
    };
  }
  );
};


$(document).ready(main);
